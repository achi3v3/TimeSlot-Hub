package backendapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"telegram-bot/pkg/models"
	mymodels "telegram-bot/pkg/models"
)

func (c *Client) RegisterUser(ctx context.Context, req mymodels.UserRegister) (string, bool) {
	url := fmt.Sprintf("%s/user/register", c.baseURL)
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.RegisterUser: %v", err)
		return err.Error(), false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("status=%d", resp.StatusCode), false
	}
	return "ok", true
}

func (c *Client) CheckAuth(ctx context.Context, telegramID int64) (string, bool) {
	url := fmt.Sprintf("%s/user/check/%d", c.baseURL, telegramID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.CheckAuth: %v", err)
		return err.Error(), false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Errorf("Adapter.BackendAPI.CheckAuth: bad status: %d, body: %s", resp.StatusCode, string(body))
		return fmt.Sprintf("server returned status %d", resp.StatusCode), false
	}

	httpResp := struct {
		Authenticated bool `json:"authenticated"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&httpResp); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.CheckAuth: decode: %v", err)
		return "failed to decode response", false
	}
	return resp.Status, httpResp.Authenticated
}

func (c *Client) ConfirmLogin(ctx context.Context, telegramID int64) (string, bool) {
	url := fmt.Sprintf("%s/user/confirm-login/%d", c.baseURL, telegramID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	secret := os.Getenv("INTERNAL_TOKEN")
	if secret == "" {
		secret = os.Getenv("TELEGRAM_HTTP_SECRET")
	}
	if secret != "" {
		httpReq.Header.Set("X-Internal-Token", secret)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.ConfirmLogin: %v", err)
		return err.Error(), false
	}
	defer resp.Body.Close()
	return resp.Status, resp.StatusCode == http.StatusOK
}

// UpdateTimezoneInternal updates user's timezone by telegram_id via internal endpoint
func (c *Client) UpdateTimezoneInternal(ctx context.Context, telegramID int64, timezone string) error {
	url := fmt.Sprintf("%s/telegram/user/timezone", c.baseURL)
	body, _ := json.Marshal(map[string]interface{}{
		"telegram_id": telegramID,
		"timezone":    timezone,
	})
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	secret := os.Getenv("INTERNAL_TOKEN")
	if secret == "" {
		secret = os.Getenv("TELEGRAM_HTTP_SECRET")
	}
	if secret != "" {
		httpReq.Header.Set("X-Internal-Token", secret)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.UpdateTimezoneInternal: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}
	return nil
}
func (c *Client) GetUserByTelegramID(ctx context.Context, userID int64) (*models.User, error) {
	url := fmt.Sprintf("%s/user/g3tter/%d", c.baseURL, userID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	secret := os.Getenv("INTERNAL_TOKEN")
	if secret == "" {
		secret = os.Getenv("TELEGRAM_HTTP_SECRET")
	}
	if secret != "" {
		httpReq.Header.Set("X-Internal-Token", secret)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetUserByTelegramID: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserByTelegramID: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserByTelegramID: %s", resp.Status)
	}
	var user *models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetUserByTelegramID: decode: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserByTelegramID: %v", err)
	}
	return user, nil
}

// ConfirmAccountDeletion подтверждает удаление аккаунта
func (c *Client) ConfirmAccountDeletion(userUUID string) error {
	url := fmt.Sprintf("%s/user/confirm-deletion", c.baseURL)
	httpReq, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)

	// Добавляем заголовки для аутентификации
	secret := os.Getenv("INTERNAL_TOKEN")
	if secret == "" {
		secret = os.Getenv("TELEGRAM_HTTP_SECRET")
	}
	if secret != "" {
		httpReq.Header.Set("X-Internal-Token", secret)
	}

	// Добавляем user_id в заголовок для идентификации пользователя
	httpReq.Header.Set("X-User-ID", userUUID)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.ConfirmAccountDeletion: %v", err)
		return fmt.Errorf("Adapter.BackendAPI.ConfirmAccountDeletion: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Adapter.BackendAPI.ConfirmAccountDeletion: status=%d", resp.StatusCode)
		return fmt.Errorf("Adapter.BackendAPI.ConfirmAccountDeletion: status=%d", resp.StatusCode)
	}

	return nil
}
