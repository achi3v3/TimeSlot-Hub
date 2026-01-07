package backendapi

import mymodels "telegram-bot/pkg/models"

type userFilterRequest struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}
type userFilterResponse struct {
	Message string            `json:"message"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	HasNext bool              `json:"has_next"`
	HasPrev bool              `json:"has_prev"`
	Records []mymodels.Record `json:"records"`
}
type updateRecordStatusRequest struct {
	RecordID uint   `json:"record_id"`
	Status   string `json:"status"`
}
