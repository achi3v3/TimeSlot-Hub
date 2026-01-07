package models

import (
	"time"

	"github.com/google/uuid"
)

// Record represents slot booking table
type Record struct {
	ID        uint      `json:"id"        gorm:"primaryKey; column:id"`
	SlotID    uint      `json:"slot_id"   gorm:"column:slot_id; not null; uniqueIndex:idx_record_slot_client"`
	ClientID  uuid.UUID `json:"client_id" gorm:"column:client_id; not null; uniqueIndex:idx_record_slot_client"`
	Status    string    `json:"status" gorm:"column:status; default:pending"`
	CreatedAt time.Time `json:"created_at" gorm:"timestamptz; column:created_at; default:CURRENT_TIMESTAMP"`

	// Expose slot in JSON so Telegram can render date/time/service/master
	Slot   Slot `json:"slot" gorm:"foreignKey:SlotID; constraint:OnDelete:CASCADE"`
	Client User `json:"client" gorm:"foreignKey:ClientID; constraint:OnDelete:CASCADE"`
}
type RecordResponce struct {
	ID        uint      `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`

	ClientID         uuid.UUID `json:"client_id"`
	ClientTelegramID int64     `json:"client_telegram_id"`
	ClientName       string    `json:"client_name"`
	ClientSurname    string    `json:"client_surname"`
	ClientPhone      string    `json:"client_phone"`

	SlotID       uint    `json:"slot_id"`
	SlotName     string  `json:"slot_name"`
	SlotPrice    float64 `json:"slot_price"`
	SlotDuration int     `json:"slot_duration"`

	MasterID         uuid.UUID `json:"master_id"`
	MasterTelegramID int64     `json:"master_telegram_id"`
	MasterName       string    `json:"master_name"`
	MasterSurname    string    `json:"master_surname"`
	MasterPhone      string    `json:"master_phone"`
}
