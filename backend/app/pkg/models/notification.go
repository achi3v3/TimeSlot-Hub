package models

// Notification types:
// - RECORD_CREATED    // "New record from client"
// - RECORD_CONFIRMED  // "Record confirmed"
// - RECORD_REJECTED   // "Record rejected"
// - SLOT_CREATED      // "Slot created"
// - SLOT_DELETED      // "Slot deleted"
// - SYSTEM_MESSAGE    // "System notification"
import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"constraint:OnDelete:CASCADE; index:idx_notification_user"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt *time.Time     `json:"expires_at"`
	Metadata  datatypes.JSON `json:"metadata"`
}
