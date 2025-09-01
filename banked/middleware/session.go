package middleware

import (
	"bvtc/log"
	"bvtc/response"
	redis_pool "bvtc/tool/pool"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionAuthMiddleware 验证session是否有效的中间件
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie("SessionId")
		if err != nil {
			log.Logger.Info("Session check failed - no cookie",
				log.String("path", c.Request.URL.Path),
				log.String("remote_addr", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, response.FailMsg("请先登录"))
			c.Abort()
			return
		}

		// 验证session是否有效
		if !validateSession(sid) {
			log.Logger.Info("Session check failed - invalid session",
				log.String("session_id", sid),
				log.String("path", c.Request.URL.Path),
				log.String("remote_addr", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, response.FailMsg("登录已过期，请重新登录"))
			c.Abort()
			return
		}

		// 将session信息存储到上下文中，供后续处理使用
		c.Set("session_id", sid)
		// log.Logger.Info("Session check passed",
		// 	log.String("session_id", sid),
		// 	log.String("path", c.Request.URL.Path),
		// 	log.String("remote_addr", c.ClientIP()))
		c.Next()
	}
}

// validateSession 验证session是否有效
func validateSession(sessionID string) bool {
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sessionID

	// 检查session是否存在
	exists, err := rdb.Exists(rtcx, key).Result()
	if err != nil {
		log.Logger.Error("Redis error checking session",
			log.String("session_id", sessionID),
			log.String("error", err.Error()))
		return false
	}
	if exists == 0 {
		log.Logger.Info("Session not found in Redis",
			log.String("session_id", sessionID))
		return false
	}

	// 检查session是否过期（Redis会自动处理过期）
	// 额外检查创建时间，确保session不会无限期有效
	createdAt, err := rdb.HGet(rtcx, key, "createdAt").Result()
	if err != nil {
		return false
	}

	// 解析创建时间 - 改进时间解析逻辑
	createTime, err := parseTime(createdAt)
	if err != nil {
		return false
	}

	// 检查session是否超过7天
	if time.Since(createTime) > 7*24*time.Hour {
		// 删除过期的session
		rdb.Del(rtcx, key)
		log.Logger.Info("Session expired and deleted",
			log.String("session_id", sessionID),
			log.String("created_at", createdAt))
		return false
	}

	// 检查session是否被标记为无效
	isValid, err := rdb.HGet(rtcx, key, "isValid").Result()
	if err == nil && isValid == "false" {
		log.Logger.Info("Session marked as invalid",
			log.String("session_id", sessionID))
		return false
	}

	// log.Logger.Info("Session validation successful",
	// 	log.String("session_id", sessionID))
	return true
}

// parseTime 改进的时间解析函数
func parseTime(timeStr string) (time.Time, error) {
	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// 尝试Unix时间戳
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// GetSessionInfo 从Redis获取session信息
func GetSessionInfo(sessionID string) (map[string]string, error) {
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sessionID

	result, err := rdb.HGetAll(rtcx, key).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}
