// Copyright (c) 2025 Youzill
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package route

import (
	"net/http"
	"os"
	"runtime"
	"strings"

	routeai "bvtc/ai"
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

	// 统一使用 /bvtc/api 路径前缀
	bvtcGroup := server.Group("/bvtc/api")
	registerRoutes(bvtcGroup)

	return server
}

// registerRoutes 注册所有API路由
func registerRoutes(group *gin.RouterGroup) {
	// 公开接口（不需要认证）
	group.POST("/netcloud/login", cloudnet.SendByPhone)          // 发送验证码
	group.POST("/netcloud/login/verify", cloudnet.VerifyCaptcha) // 验证验证码

	// 测试接口
	group.GET("/test/bilibili/download", bilibili.DownloadVideo)
	group.GET("/test/bilibili/desc", bilibili.GetVideoDesc)

	// 需要认证的接口
	authGroup := group.Group("/")
	authGroup.Use(middleware.SessionAuthMiddleware())
	{
		authGroup.GET("/netcloud/login/check", cloudnet.CheckCookie)  // 检查登陆状态
		authGroup.POST("/netcloud/logout", cloudnet.DeleteCookie)     // 退出登录,删除状态（改为POST防CSRF）
		authGroup.GET("/netcloud/playlist", cloudnet.ShowPlaylist)    //获取歌单
		authGroup.GET("/netcloud/useravatar", cloudnet.GetUserAvatar) //获取用户头像

		authGroup.POST("/bilibili/createtask", bilibili.CreateLoadMP4Task)                     // 创建任务
		authGroup.GET("/bilibili/checktask/:taskId", bilibili.CheckLoadMP4Task)                // 查询任务状态
		authGroup.GET("/bilibili/list", bilibili.GetVideoList)                                 // 视频列表
		authGroup.GET("/bilibili/suggest-title-batch/stream", routeai.SuggestTitleBatchStream) // 生成标题（SSE流式）
		// 暂时不用下面接口
		authGroup.GET("/bilibili/login", bilibili.BiliLogin)
		authGroup.GET("/bilibili/login/check", bilibili.BiliLoginWithCookie)
	}
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
