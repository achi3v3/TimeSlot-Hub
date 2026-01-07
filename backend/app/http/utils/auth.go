package utils

import (
	"app/encoder"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// ExtractUserIDFromToken извлекает user_id из Bearer токена
func ExtractUserIDFromToken(ctx *gin.Context) (uuid.UUID, error) {
	auth := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return uuid.Nil, fmt.Errorf("missing bearer token")
	}
	tokenString := strings.TrimPrefix(auth, "Bearer ")
	token, err := encoder.ParseToken(tokenString)
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}
	raw := claims["user_id"]
	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("user_id not found in token")
	}
}

// ParseToken парсит JWT токен
func ParseToken(tokenString string) (*jwt.Token, error) {
	return encoder.ParseToken(tokenString)
}
