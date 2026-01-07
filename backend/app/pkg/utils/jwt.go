package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTSecret = []byte("your-secret-key-change-in-production") // В продакшене должен быть в переменных окружения
	JWTExpiry = 24 * time.Hour
)

// Claims структура для JWT claims
type Claims struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
	Phone      string `json:"phone"`
	FirstName  string `json:"first_name"`
	Surname    string `json:"surname"`
	jwt.RegisteredClaims
}

// GenerateJWT создает JWT токен для пользователя
func GenerateJWT(userID string, telegramID int64, phone, firstName, surname string) (string, error) {
	claims := Claims{
		UserID:     userID,
		TelegramID: telegramID,
		Phone:      phone,
		FirstName:  firstName,
		Surname:    surname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWTExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "botans-app",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// ValidateJWT валидирует JWT токен и возвращает claims
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем срок действия
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, errors.New("token has expired")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshJWT обновляет JWT токен
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Извлекаем данные из claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid user_id in token")
	}

	telegramID, ok := claims["telegram_id"].(float64)
	if !ok {
		return "", errors.New("invalid telegram_id in token")
	}

	phone, ok := claims["phone"].(string)
	if !ok {
		phone = ""
	}

	firstName, ok := claims["first_name"].(string)
	if !ok {
		firstName = ""
	}

	surname, ok := claims["surname"].(string)
	if !ok {
		surname = ""
	}

	// Генерируем новый токен
	return GenerateJWT(userID, int64(telegramID), phone, firstName, surname)
}
