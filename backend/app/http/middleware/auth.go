package middleware

import (
	"app/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware проверяет JWT токен и устанавливает пользователя в контекст
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Валидируем JWT токен
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Извлекаем user_id из claims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Устанавливаем user_id в контекст для использования в хендлерах
		c.Set("user_id", userID)
		c.Set("user_telegram_id", claims["telegram_id"])

		c.Next()
	}
}

// OptionalAuthMiddleware проверяет токен если он есть, но не требует его обязательного наличия
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.Next()
			return
		}

		if userIDStr, ok := claims["user_id"].(string); ok {
			if userID, err := uuid.Parse(userIDStr); err == nil {
				c.Set("user_id", userID)
				c.Set("user_telegram_id", claims["telegram_id"])
			}
		}

		c.Next()
	}
}

// RequireRoleMiddleware проверяет, что у пользователя есть определенная роль
func RequireRoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Здесь нужно проверить роль пользователя
		// Для простоты пока что пропускаем всех аутентифицированных пользователей
		// В реальном приложении нужно добавить проверку ролей в базе данных

		c.Next()
	}
}

// ResourceOwnerMiddleware проверяет, что пользователь является владельцем ресурса
func ResourceOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Проверяем, что user_id в запросе совпадает с аутентифицированным пользователем
		// Это нужно для эндпоинтов типа /user/update, где user_id передается в теле запроса
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				if requestUserID, ok := requestBody["user_id"].(string); ok {
					if requestUserID != userID.(uuid.UUID).String() {
						c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: you can only modify your own resources"})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// MasterOnlyMiddleware проверяет, что пользователь является мастером
func MasterOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Здесь нужно проверить, что пользователь является мастером
		// Пока что пропускаем всех аутентифицированных пользователей
		// В реальном приложении нужно добавить проверку роли "master" в базе данных

		c.Next()
	}
}

// RateLimitMiddleware ограничивает количество запросов
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Простая реализация rate limiting
		// В продакшене лучше использовать Redis или специализированные библиотеки
		_ = c.ClientIP()

		// Здесь должна быть логика проверки лимитов по IP
		// Пока что пропускаем все запросы

		c.Next()
	}
}

// ValidateInputMiddleware валидирует входные данные
func ValidateInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем Content-Type только если у запроса есть тело
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// Если тело отсутствует (Content-Length == 0), не навязываем application/json
			if c.Request.ContentLength > 0 {
				contentType := c.GetHeader("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
					c.Abort()
					return
				}
			}
		}

		// Ограничиваем размер тела запроса
		if c.Request.ContentLength > 1024*1024 { // 1MB
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large"})
			c.Abort()
			return
		}

		// Проверяем длину URL
		if len(c.Request.URL.Path) > 2048 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL too long"})
			c.Abort()
			return
		}

		// Проверяем заголовки на подозрительное содержимое
		for _, values := range c.Request.Header {
			for _, value := range values {
				if len(value) > 1024 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Header value too long"})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
