package middleware

import (
	"app/http/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// AdminPhoneMiddleware проверяет, что запрос от админа по номеру телефона
func AdminPhoneMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем номер телефона из заголовка или параметра
		phone := c.GetHeader("X-Admin-Phone")
		if phone == "" {
			phone = c.Query("admin_phone")
		}

		// Проверяем, что это номер админа
		if phone != "79876038494" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: admin phone required"})
			c.Abort()
			return
		}

		// Устанавливаем флаг админа в контекст
		c.Set("is_admin", true)
		c.Set("admin_phone", phone)

		c.Next()
	}
}

// AdminAuthMiddleware проверяет авторизацию админа через Telegram
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что пользователь аутентифицирован
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Получаем telegram_id пользователя
		telegramID, exists := c.Get("user_telegram_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Telegram ID not found"})
			c.Abort()
			return
		}

		// Проверяем, что это админ (по telegram_id или по номеру телефона)
		// Для MVP проверяем по telegram_id
		adminTelegramID := int64(123456789) // Замените на ваш telegram_id

		if telegramID != adminTelegramID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: admin privileges required"})
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// AdminOrPhoneMiddleware проверяет либо админскую авторизацию, либо номер телефона
func AdminOrPhoneMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Сначала проверяем авторизацию через JWT токен
		auth := c.GetHeader("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			tokenString := strings.TrimPrefix(auth, "Bearer ")
			token, err := utils.ParseToken(tokenString)
			if err == nil && token.Valid {
				claims, ok := token.Claims.(jwt.MapClaims)
				if ok {
					userIDStr, ok := claims["user_id"].(string)
					if ok {
						userID, err := uuid.Parse(userIDStr)
						if err == nil {
							// Пользователь аутентифицирован через JWT
							// Для админ-панели: любой аутентифицированный пользователь считается админом
							c.Set("user_id", userID)
							c.Set("is_admin", true)
							c.Next()
							return
						}
					}
				}
			}
		}

		// Если JWT авторизация не прошла, проверяем номер телефона (для обратной совместимости)
		phone := c.GetHeader("X-Admin-Phone")
		if phone == "" {
			phone = c.Query("admin_phone")
		}

		if phone == "79876038494" {
			c.Set("is_admin", true)
			c.Set("admin_phone", phone)
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: admin authentication required"})
		c.Abort()
	}
}
