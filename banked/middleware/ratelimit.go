package middleware

import (
	"bvtc/response"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.limiters[ip] = limiter
	}

	return limiter
}

func RateLimitMiddleware(requestsPerMinute int, burstSize int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(float64(requestsPerMinute)/60.0), burstSize)
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.AddIP(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, response.FailMsg("请求频率过高，请稍后再试"))
			c.Abort()
			return
		}
		c.Next()
	}
}
