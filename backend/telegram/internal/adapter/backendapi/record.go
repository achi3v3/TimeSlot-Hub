package backendapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	mymodels "telegram-bot/pkg/models"
)

func (c *Client) CreateRecord(ctx context.Context, req mymodels.Record) (string, bool) {
	url := fmt.Sprintf("%s/telegram/record/master/create", c.baseURL)
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
		c.logger.Errorf("Adapter.BackendAPI.RegisterUser: %v", err)
		return err.Error(), false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("status=%d", resp.StatusCode), false
	}
	return "ok", true
}

func (c *Client) GetUserRecordsFiltered(ctx context.Context, telegramID int64, status string, page, limit int) (*userFilterResponse, error) {
	user, err := c.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("GetUserRecordsFiltered: %v", err)
	}
	url := fmt.Sprintf("%s/record/user/filter", c.baseURL)
	payload := userFilterRequest{UserID: user.ID.String(), Status: status, Page: page, Limit: limit}
	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetUserRecordsFiltered: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserRecordsFiltered: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserRecordsFiltered: status=%d", resp.StatusCode)
	}
	var out userFilterResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetUserRecordsFiltered: decode: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetUserRecordsFiltered: %v", err)
	}
	return &out, nil
}
func (c *Client) GetRecords(ctx context.Context, userID int64) ([]mymodels.Record, error) {
	user, err := c.GetUserByTelegramID(ctx, userID)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
	}
	url := fmt.Sprintf("%s/record/%s", c.baseURL, user.ID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
	}
	var records []mymodels.Record
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetRecords: decode: %v", err)
		return nil, fmt.Errorf("Adapter.BackendAPI.GetRecords: %v", err)
	}
	return records, nil
}

func (c *Client) UpdateRecordStatus(ctx context.Context, recordID uint, status string) (string, bool) {
	url := fmt.Sprintf("%s/telegram/record/master/status", c.baseURL)
	payload := updateRecordStatusRequest{RecordID: recordID, Status: status}
	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
		c.logger.Errorf("Adapter.BackendAPI.UpdateRecordStatus: %v", err)
		return err.Error(), false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("status=%d", resp.StatusCode), false
	}
	return "ok", true
}

// GetUpcomingRecordsByMasterTelegramID получает предстоящие записи мастера
func (c *Client) GetUpcomingRecordsByMasterTelegramID(ctx context.Context, masterTelegramID int64) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/telegram/record/master/upcoming/%d", c.baseURL, masterTelegramID)
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
		c.logger.Errorf("Adapter.BackendAPI.GetUpcomingRecordsByMasterTelegramID: %v", err)
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("Adapter.BackendAPI.GetUpcomingRecordsByMasterTelegramID: status=%d", resp.StatusCode)
		return nil, fmt.Errorf("status=%d", resp.StatusCode)
	}

	var result struct {
		Message string                   `json:"message"`
		Data    []map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetUpcomingRecordsByMasterTelegramID: decode: %v", err)
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return result.Data, nil
}
