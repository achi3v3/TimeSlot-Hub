package middleware

import (
	"app/http/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// SessionAuthMiddleware проверяет сессионный токен и устанавливает пользователя в контекст
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ExtractUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Устанавливаем user_id в контекст для использования в хендлерах
		c.Set("user_id", userID)
		c.Next()
	}
}

// SessionAuthMiddlewareWithTelegramID проверяет сессионный токен и устанавливает пользователя с telegram_id в контекст
func SessionAuthMiddlewareWithTelegramID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ExtractUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Извлекаем telegram_id из токена
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Парсим токен для получения telegram_id
		tokenString := auth[7:] // Убираем "Bearer "
		token, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		telegramID, ok := claims["telegram_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Telegram ID not found in token"})
			c.Abort()
			return
		}

		// Устанавливаем данные в контекст
		c.Set("user_id", userID)
		c.Set("user_telegram_id", int64(telegramID))
		c.Next()
	}
}
