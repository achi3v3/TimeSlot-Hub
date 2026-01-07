package middleware

import (
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizeInputMiddleware санитизирует входные данные
func SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем запросы, которые не содержат JSON
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		// Читаем тело запроса
		body, err := c.GetRawData()
		if err != nil {
			c.Next()
			return
		}

		// Санитизируем JSON
		sanitizedBody := sanitizeJSON(string(body))

		// Устанавливаем санитизированное тело обратно
		c.Request.Body = io.NopCloser(strings.NewReader(sanitizedBody))

		c.Next()
	}
}

// sanitizeJSON санитизирует JSON строку
func sanitizeJSON(jsonStr string) string {
	// Удаляем потенциально опасные символы
	jsonStr = strings.ReplaceAll(jsonStr, "<script", "&lt;script")
	jsonStr = strings.ReplaceAll(jsonStr, "</script>", "&lt;/script&gt;")
	jsonStr = strings.ReplaceAll(jsonStr, "javascript:", "")
	jsonStr = strings.ReplaceAll(jsonStr, "vbscript:", "")
	jsonStr = strings.ReplaceAll(jsonStr, "onload=", "")
	jsonStr = strings.ReplaceAll(jsonStr, "onerror=", "")
	jsonStr = strings.ReplaceAll(jsonStr, "onclick=", "")

	// Экранируем HTML сущности
	jsonStr = html.EscapeString(jsonStr)

	return jsonStr
}

// SanitizeString санитизирует строку
func SanitizeString(input string) string {
	// Удаляем HTML теги
	re := regexp.MustCompile(`<[^>]*>`)
	input = re.ReplaceAllString(input, "")

	// Экранируем специальные символы
	input = html.EscapeString(input)

	// Удаляем потенциально опасные паттерны
	dangerousPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
	}

	for _, pattern := range dangerousPatterns {
		input = strings.ReplaceAll(input, pattern, "")
	}

	return strings.TrimSpace(input)
}

// ValidatePhoneNumber валидирует номер телефона
func ValidatePhoneNumber(phone string) bool {
	// Удаляем все нецифровые символы
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(phone, "")

	// Проверяем длину (от 10 до 15 цифр)
	return len(digits) >= 10 && len(digits) <= 15
}

// ValidateEmail валидирует email
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// ValidateUUID валидирует UUID
func ValidateUUID(uuid string) bool {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return re.MatchString(strings.ToLower(uuid))
}
