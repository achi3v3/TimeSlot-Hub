package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter структура для rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// IsAllowed проверяет, разрешен ли запрос
func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Очищаем старые запросы
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// Проверяем лимит
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Добавляем новый запрос
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// Глобальные rate limiters для разных типов запросов
var (
	// Общий rate limiter: 100 запросов в минуту
	generalLimiter = NewRateLimiter(100, time.Minute)

	// Строгий rate limiter для аутентификации: 5 попыток в минуту
	authLimiter = NewRateLimiter(5, time.Minute)

	// Rate limiter для создания ресурсов: 10 запросов в минуту
	createLimiter = NewRateLimiter(10, time.Minute)
)

// GeneralRateLimitMiddleware общий rate limiting
func GeneralRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !generalLimiter.IsAllowed(clientIP) {
			c.JSON(429, gin.H{
				"error":       "Too many requests",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware строгий rate limiting для аутентификации
func AuthRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !authLimiter.IsAllowed(clientIP) {
			c.JSON(429, gin.H{
				"error":       "Too many authentication attempts",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateRateLimitMiddleware rate limiting для создания ресурсов
func CreateRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !createLimiter.IsAllowed(clientIP) {
			c.JSON(429, gin.H{
				"error":       "Too many creation requests",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimitMiddleware rate limiting по пользователю
func UserRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// Если пользователь не аутентифицирован, используем IP
			clientIP := c.ClientIP()
			if !generalLimiter.IsAllowed(clientIP) {
				c.JSON(429, gin.H{
					"error":       "Too many requests",
					"retry_after": 60,
				})
				c.Abort()
				return
			}
		} else {
			// Если пользователь аутентифицирован, используем его ID
			userKey := "user_" + userID.(string)
			if !generalLimiter.IsAllowed(userKey) {
				c.JSON(429, gin.H{
					"error":       "Too many requests",
					"retry_after": 60,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
