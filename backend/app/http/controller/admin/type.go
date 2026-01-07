package admin

import (
	"app/http/repository/metrics"
	"app/http/repository/record"
	"app/http/repository/service"
	"app/http/repository/slot"
	"app/http/repository/user"
	recordServ "app/http/usecase/record"
	serviceServ "app/http/usecase/service"
	slotServ "app/http/usecase/slot"
	userServ "app/http/usecase/user"
	"app/pkg/models"
	"time"
)

type Handler struct {
	userRepo    *user.Repository
	slotRepo    *slot.Repository
	serviceRepo *service.Repository
	recordRepo  *record.Repository
	metricsRepo *metrics.Repository

	userServ    *userServ.Service
	slotServ    *slotServ.Service
	recordServ  *recordServ.Service
	serviceServ *serviceServ.Service

	logger Logger
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func NewHandler(
	userRepo *user.Repository,
	slotRepo *slot.Repository,
	serviceRepo *service.Repository,
	recordRepo *record.Repository,
	metricsRepo *metrics.Repository,
	userServ *userServ.Service,
	slotServ *slotServ.Service,
	serviceServ *serviceServ.Service,
	recordServ *recordServ.Service,
	logger Logger,
) *Handler {
	return &Handler{
		userRepo:    userRepo,
		slotRepo:    slotRepo,
		serviceRepo: serviceRepo,
		recordRepo:  recordRepo,
		metricsRepo: metricsRepo,
		userServ:    userServ,
		slotServ:    slotServ,
		serviceServ: serviceServ,
		recordServ:  recordServ,
		logger:      logger,
	}
}

// AdminStatsResponse структура для статистики
type AdminStatsResponse struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	TotalSlots       int64 `json:"total_slots"`
	BookedSlots      int64 `json:"booked_slots"`
	TotalRecords     int64 `json:"total_records"`
	PendingRecords   int64 `json:"pending_records"`
	ConfirmedRecords int64 `json:"confirmed_records"`
	RejectedRecords  int64 `json:"rejected_records"`
	TotalServices    int64 `json:"total_services"`
	AdClicks1        int64 `json:"ad_clicks_1"`
	AdClicks2        int64 `json:"ad_clicks_2"`
}

// UserDetailResponse структура для детальной информации о пользователе
type UserDetailResponse struct {
	User     UserInfo         `json:"user"`
	Slots    []models.Slot    `json:"slots"`
	Services []models.Service `json:"services"`
	Records  []models.Record  `json:"records"`
}

// AdminLoginRequest запрос на авторизацию админа
type AdminLoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse ответ на авторизацию админа
type AdminLoginResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// UserListResponse структура для списка пользователей
type UserListResponse struct {
	Users []UserInfo `json:"users"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// UserInfo структура для информации о пользователе
type UserInfo struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"first_name"`
	Surname    string    `json:"surname"`
	Phone      string    `json:"phone"`
	TelegramID int64     `json:"telegram_id"`
	IsActive   bool      `json:"is_active"`
	Roles      []string  `json:"roles"`
	CreatedAt  time.Time `json:"created_at"`
}
