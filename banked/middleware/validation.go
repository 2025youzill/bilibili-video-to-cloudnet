package middleware

import (
	"bvtc/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求大小
		if c.Request.ContentLength > 10*1024*1024 { // 10MB
			c.JSON(http.StatusRequestEntityTooLarge, response.FailMsg("请求数据过大"))
			c.Abort()
			return
		}

		// 检查Content-Type
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// 只检查Content-Type，不消费请求体
			// 实际的JSON验证由具体的处理函数完成
		}

		c.Next()
	}
}
