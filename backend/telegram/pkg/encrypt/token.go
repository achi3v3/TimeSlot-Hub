package encrypt

import (
	"telegram-bot/internal/config"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte(config.GetEnv("JWT_SECRET", "example-secret-key"))

// GenerateToken возвращает JWT-токен по user_id и secretKey, исключение
func GenerateToken(userID int64, phone string) (string, error) {
	claims := jwt.MapClaims{
		"phone":   phone,
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseToken расшифровывает JWT-токен, возвращает токен и исключение
func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}
