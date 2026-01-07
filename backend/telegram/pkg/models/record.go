package models

import (
	"time"

	"github.com/google/uuid"
)

// Таблица бронирования слотов
type Record struct {
	ID       uint      `json:"id"`
	SlotID   uint      `json:"slot_id"`
	ClientID uuid.UUID `json:"client_id"`
	Status   string    `json:"status"`

	Slot   Slot `json:"slot"`
	Client User `json:"client"`
}
type Slot struct {
	ID        uint      `json:"id"          gorm:"primaryKey; column:id"`
	MasterID  uuid.UUID `json:"master_id"   gorm:"column:master_id; not null"`
	StartTime time.Time `json:"start_time"  gorm:"column:start_time"`
	EndTime   time.Time `json:"end_time"    gorm:"column:end_time"`
	IsBooked  bool      `json:"is_booked"   gorm:"column:is_booked; default:false"`
	ServiceID uint      `json:"service_id"  gorm:"column:service_id; not null"`

	Service Service `json:"service" gorm:"foreignKey:ServiceID; constraint:OnDelete:CASCADE"`
	Master  User    `json:"master" gorm:"foreignKey:MasterID; constraint:OnDelete:CASCADE"`
}
