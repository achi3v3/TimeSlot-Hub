package models

import (
	"github.com/google/uuid"
)

type Service struct {
	ID          uint      `json:"id"`
	MasterID    uuid.UUID `json:"master_id" gorm:"index:idx_service_master"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
}
type ServiceResponse struct {
	ID          uint      `json:"id"`
	MasterID    uuid.UUID `json:"master_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`

	MasterTelegramID int64  `json:"master_telegram_id"`
	MasterName       string `json:"master_name"`
	MasterSurname    string `json:"master_surname"`
	MasterPhone      string `json:"master_phone"`
}
