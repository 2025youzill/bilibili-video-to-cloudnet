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
