package user

import (
	"app/http/repository/user"

	"github.com/sirupsen/logrus"
)

type Service struct {
	repo   *user.Repository
	logger *logrus.Logger
}

// NewUserService - конструктор Service.
// Ответ: Возвращает ссылку на структуру Service.
func NewService(repo *user.Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// UpdateNamesRequest описывает разрешенные для изменения поля пользователя
type UpdateNamesRequest struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	Surname   string `json:"surname"`
}

type UpdateTimezoneRequest struct {
	UserID   string `json:"user_id"`
	Timezone string `json:"timezone"`
}
