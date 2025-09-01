package route

import (
	"net/http"
	"os"
	"runtime"
	"strings"

	"bvtc/bilibili"
	"bvtc/cloudnet"
	"bvtc/config"
	"bvtc/log"
	"bvtc/middleware"
	"bvtc/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()

	// 基础中间件
	server.Use(Cors())
	server.Use(Recovery)
	server.Use(middleware.SecurityHeadersMiddleware())
	server.Use(middleware.InputValidationMiddleware())

	// 限流中间件（每分钟60请求，突发10个）
	server.Use(middleware.RateLimitMiddleware(60, 10))

	// 健康检查放在根路径
	server.GET("/health", HealthCheck)

	group := server.Group("/api")
	{
		// 公开接口（不需要认证）
		group.POST("/netcloud/login", cloudnet.SendByPhone)          // 发送验证码
		group.POST("/netcloud/login/verify", cloudnet.VerifyCaptcha) // 验证验证码

		// 需要认证的接口
		authGroup := group.Group("/")
		authGroup.Use(middleware.SessionAuthMiddleware())
		{
			authGroup.GET("/netcloud/login/check", cloudnet.CheckCookie)  // 检查登陆状态
			authGroup.POST("/netcloud/logout", cloudnet.DeleteCookie)     // 退出登录,删除状态（改为POST防CSRF）
			authGroup.GET("/netcloud/playlist", cloudnet.ShowPlaylist)    //获取歌单
			authGroup.GET("/netcloud/useravatar", cloudnet.GetUserAvatar) //获取用户头像

			authGroup.POST("/bilibili/createtask", bilibili.CreateLoadMP4Task)      // 创建任务
			authGroup.GET("/bilibili/checktask/:taskId", bilibili.CheckLoadMP4Task) // 查询任务状态
			authGroup.GET("/bilibili/list", bilibili.GetVideoList)

			// 暂时不用下面接口
			authGroup.GET("/bilibili/login", bilibili.BiliLogin)
			authGroup.GET("/bilibili/login/check", bilibili.BiliLoginWithCookie)
		}
	}
	return server
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 直接从环境变量读取
		envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if envOrigins == "" {
			panic("CORS_ALLOWED_ORIGINS environment variable is not set")
		}
		allowedOrigins := strings.Split(envOrigins, ",")
		allowed := false

		// // 调试日志
		// log.Logger.Info("CORS check",
		// 	log.String("origin", origin),
		// 	log.String("allowedOrigins", allowedOrigins))

		for _, allowedOrigin := range allowedOrigins {
			allowedOrigin = strings.TrimSpace(allowedOrigin)
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// 生产环境应该限制为特定域名
		if !allowed && origin != "" {
			log.Logger.Warn("Origin not allowed",
				log.String("origin", origin),
				log.String("allowedOrigins", strings.Join(allowedOrigins, ", ")))
			c.JSON(http.StatusForbidden, response.FailMsg("Origin not allowed"))
			c.Abort()
			return
		}

		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)

			// 使用配置文件中的方法，如果没有配置则使用默认值
			allowedMethods := config.GetConfig().Security.CORS.AllowedMethods
			if len(allowedMethods) == 0 {
				allowedMethods = []string{"POST", "GET", "OPTIONS"}
			}
			c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))

			// 使用配置文件中的头部，如果没有配置则使用默认值
			allowedHeaders := config.GetConfig().Security.CORS.AllowedHeaders
			if len(allowedHeaders) == 0 {
				allowedHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
			}
			c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "3600") // 1小时
		}

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		defer func() {
			if err := recover(); err != nil {
				log.Logger.Error("HttpError", zap.Any("HttpError", err))
			}
		}()

		c.Next()
	}
}

func Recovery(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			log.Logger.Error("gin catch panic: ",
				log.Any("error", r),
				log.String("stack", string(stack[:length])),
			)
			c.JSON(http.StatusOK, response.FailMsg("系统内部错误"))
		}
	}()
	c.Next()
}

// HealthCheck 健康检查处理函数
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessMsg("Server is healthy"))
}
