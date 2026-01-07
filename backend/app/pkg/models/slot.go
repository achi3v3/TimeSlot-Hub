package models

import (
	"time"

	"github.com/google/uuid"
)

// Slot represents master slots table
type Slot struct {
	ID        uint      `json:"id"          gorm:"primaryKey; column:id"`
	MasterID  uuid.UUID `json:"master_id"   gorm:"column:master_id; not null; ; index:idx_slot_master_time"`
	StartTime time.Time `json:"start_time"  gorm:"column:start_time; index:idx_slot_master_time"`
	EndTime   time.Time `json:"end_time"    gorm:"column:end_time"`
	IsBooked  bool      `json:"is_booked"   gorm:"column:is_booked; default:false"`
	ServiceID uint      `json:"service_id"  gorm:"column:service_id; not null"`

	Service Service `json:"service" gorm:"foreignKey:ServiceID; constraint:OnDelete:CASCADE"`
	Master  User    `json:"master" gorm:"foreignKey:MasterID; constraint:OnDelete:CASCADE"`
}
type SlotResponse struct {
	ID        uint      `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsBooked  bool      `json:"is_booked"`

	ServiceName        string  `json:"service_name"`
	ServiceDescription string  `json:"service_description"`
	ServicePrice       float64 `json:"service_price"`
	ServiceDuration    int     `json:"service_duration"`

	MasterTelegramID int64  `json:"master_telegram_id"`
	MasterName       string `json:"master_name"`
	MasterSurname    string `json:"master_surname"`
	MasterPhone      string `json:"master_phone"`
}
