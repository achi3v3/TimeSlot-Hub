package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents users table
type User struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone                   string    `json:"phone" gorm:"unique; not null; column:phone"`
	TelegramID              int64     `json:"telegram_id" gorm:"uniqueIndex; column:telegram_id"`
	FirstName               string    `json:"first_name" gorm:"column:first_name; not null"`
	Surname                 string    `json:"surname" gorm:"column:surname; not null"`
	Timezone                string    `json:"timezone" gorm:"column:timezone; default:'Europe/Moscow'"`
	Active                  bool      `json:"active" gorm:"column:active; default:false"`
	ConsentGivenAt          time.Time `json:"consent_given_at" gorm:"timestamptz; column:consent_given_at"`
	PrivacyPolicyAcceptedAt time.Time `json:"privacy_policy_accepted_at" gorm:"timestamptz; column:privacy_policy_accepted_at"`
	TermsAcceptedAt         time.Time `json:"terms_accepted_at" gorm:"timestamptz; column:terms_accepted_at"`

	Roles    []UserRole `json:"roles"       gorm:"foreignKey:UserID; default:'[]'; constraint:OnDelete:CASCADE"`
	Services []Service  `json:"services"    gorm:"foreignKey:MasterID; default:'[]'; constraint:OnDelete:CASCADE"`
}

// UserRole represents user roles table
type UserRole struct {
	UserID uuid.UUID `json:"user_id" gorm:"primaryKey; index; not null; constraint:OnDelete:CASCADE"`
	Role   string    `json:"role"    gorm:"primaryKey"`
}
