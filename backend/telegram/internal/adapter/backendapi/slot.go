package backendapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"telegram-bot/pkg/models"
)

func (c *Client) GetSlotsByTelegramID(ctx context.Context, masterID int64) ([]models.SlotResponse, bool) {
	user, err := c.GetUserByTelegramID(ctx, masterID)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetSlotsByTelegramID: %v", err)
		return nil, false
	}
	url := fmt.Sprintf("%s/slot/%s", c.baseURL, user.ID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetSlotsByTelegramID: %v", err)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	var slots []models.SlotResponse
	if err := json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetSlotsByTelegramID: decode: %v", err)
		return nil, false
	}
	return slots, true
}

func (c *Client) GetSlotByID(ctx context.Context, slotID uint) (*models.SlotResponse, bool) {
	url := fmt.Sprintf("%s/slot/one/%d", c.baseURL, slotID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetSlotsByTelegramID: %v", err)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	var slot *models.SlotResponse
	if err := json.NewDecoder(resp.Body).Decode(&slot); err != nil {
		c.logger.Errorf("Adapter.BackendAPI.GetSlotsByTelegramID: decode: %v", err)
		return nil, false
	}
	return slot, true
}

func (c *Client) DeleteSlotsByTelegramID(ctx context.Context, masterID uint) bool {
	url := fmt.Sprintf("%s/slot/master/%d", c.baseURL, masterID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Errorf("Adapter.BackendAPI.DeleteSlotsByTelegramID: %v", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
