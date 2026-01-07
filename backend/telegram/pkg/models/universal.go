package models

import "time"

type UserRegister struct {
	Phone      string `json:"phone"`
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	Surname    string `json:"surname"`
	Token      string `json:"token"`
	Active     bool   `json:"active"`
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
	MasterTimezone   string `json:"master_timezone"`
}
