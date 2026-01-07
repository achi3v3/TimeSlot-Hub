package sender

import (
	"app/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// AuthNotify отправляет запрос в сервис telegram-bot для подтверждения входа
// Ответ: Возвращает ошибку
func LoginNotify(user models.User, ip string, location string) error {
	base := os.Getenv("TELEGRAM_HTTP_BASE")
	if base == "" {
		base = "http://telegram:8091"
	}
	url := fmt.Sprintf("%s/notify-login/%d", base, user.TelegramID)
	// pass meta via query params (best-effort)
	if ip != "" || location != "" {
		q := "?"
		if ip != "" {
			q += "ip=" + ip
		}
		if location != "" {
			if q != "?" {
				q += "&"
			}
			q += "loc=" + location
		}
		url += q
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %v", err)
	}
	// Добавляем общий внутренний токен для защиты HTTP-нотификатора
	// Use TELEGRAM_HTTP_SECRET if provided, fallback to INTERNAL_TOKEN for compatibility
	if tok := func() string {
		if v := os.Getenv("TELEGRAM_HTTP_SECRET"); v != "" {
			return v
		}
		return os.Getenv("INTERNAL_TOKEN")
	}(); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}
	responce, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки уведомления: %v", err)
	}
	defer responce.Body.Close()

	if responce.StatusCode != http.StatusOK {
		return fmt.Errorf("Сервер вернул статус: %d", responce.StatusCode)
	}

	return nil
}

// RecordNotify отправляет POST-запрос на сервис telegram-bot для отправки произвольного уведомления
// телеграм-пользователю. Используется для уведомлений о записи/статусах.
func RecordNotify(recordID uint, telegramID int64, title, message string) error {
	base := os.Getenv("TELEGRAM_HTTP_BASE")
	if base == "" {
		base = "http://telegram:8091"
	}
	url := fmt.Sprintf("%s/notify-record", base)

	payload := struct {
		RecordID   uint   `json:"record_id"`
		TelegramID int64  `json:"telegram_id"`
		Title      string `json:"title"`
		Message    string `json:"message"`
	}{
		RecordID:   recordID,
		TelegramID: telegramID,
		Title:      title,
		Message:    message,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации тела запроса: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := func() string {
		if v := os.Getenv("TELEGRAM_HTTP_SECRET"); v != "" {
			return v
		}
		return os.Getenv("INTERNAL_TOKEN")
	}(); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус: %d", resp.StatusCode)
	}
	return nil
}

// RecordStatusNotify отправляет POST-запрос на сервис telegram-bot для отправки уведомления о статусе записи
func RecordStatusNotify(telegramID int64, title, message string) error {
	base := os.Getenv("TELEGRAM_HTTP_BASE")
	if base == "" {
		base = "http://telegram:8091"
	}
	url := fmt.Sprintf("%s/notify-record-status", base)

	payload := struct {
		TelegramID int64  `json:"telegram_id"`
		Title      string `json:"title"`
		Message    string `json:"message"`
	}{
		TelegramID: telegramID,
		Title:      title,
		Message:    message,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации тела запроса: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := func() string {
		if v := os.Getenv("TELEGRAM_HTTP_SECRET"); v != "" {
			return v
		}
		return os.Getenv("INTERNAL_TOKEN")
	}(); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус: %d", resp.StatusCode)
	}
	return nil
}

// RequestAccountDeletionConfirmation отправляет запрос на подтверждение удаления аккаунта в Telegram
func RequestAccountDeletionConfirmation(userID uuid.UUID, telegramID int64) error {
	base := os.Getenv("TELEGRAM_HTTP_BASE")
	if base == "" {
		base = "http://telegram:8091"
	}
	url := fmt.Sprintf("%s/notify-account-deletion", base)

	payload := struct {
		UserID     string `json:"user_id"`
		TelegramID int64  `json:"telegram_id"`
	}{
		UserID:     userID.String(),
		TelegramID: telegramID,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации тела запроса: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := func() string {
		if v := os.Getenv("TELEGRAM_HTTP_SECRET"); v != "" {
			return v
		}
		return os.Getenv("INTERNAL_TOKEN")
	}(); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус: %d", resp.StatusCode)
	}
	return nil
}
