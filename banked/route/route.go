package route

import (
	"net/http"
	"runtime"

	"bvtc/bilibili"
	"bvtc/cloudnet"
	"bvtc/log"
	"bvtc/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	server.Use(Cors())
	server.Use(Recovery)

	group := server.Group("")
	{
		group.POST("/netcloud/login", cloudnet.SendByPhone)          // 发送验证码
		group.POST("/netcloud/login/verify", cloudnet.VerifyCaptcha) // 验证验证码
		group.GET("/netcloud/login/check", cloudnet.CheckCookie)     // 检查登陆状态
		group.GET("/netcloud/playlist", cloudnet.ShowPlaylist)

		group.POST("/bilibili/load", bilibili.LoadMP4)
		group.GET("/bilibili/list", bilibili.GetVideoList)

		// 暂时不用下面接口
		group.GET("/bilibili/login", bilibili.BiliLogin)
		group.GET("/bilibili/login/check", bilibili.BiliLoginWithCookie)
	}
	return server
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// 修改此处：将*替换为前端实际域名（如http://localhost:3000）
			c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true") // 保持凭证允许
			c.Header("Access-Control-Max-Age", "86400")          // 预检请求结果缓存24小时
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
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
