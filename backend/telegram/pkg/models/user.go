package models

import "github.com/google/uuid"

// User represents user table
type User struct {
	ID         uuid.UUID `json:"id"`
	Phone      string    `json:"phone"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	Surname    string    `json:"surname"`
	Timezone   string    `json:"timezone"`
	Active     bool      `json:"active"`

	Roles    []UserRole `json:"roles"`
	Services []Service  `json:"services"`
}

// UserRole represents user roles table
type UserRole struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}
type Service struct {
	ID          uint      `json:"id"`
	MasterID    uuid.UUID `json:"master_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
}
