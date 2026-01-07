package login

import (
	"context"
	adapter "telegram-bot/internal/adapter/backendapi"
	mymodels "telegram-bot/pkg/models"

	"github.com/sirupsen/logrus"
)

type Service struct {
	api    *adapter.Client
	logger *logrus.Logger
}

func New(api *adapter.Client, logger *logrus.Logger) *Service {
	return &Service{api: api, logger: logger}
}

func (s *Service) RegisterUser(ctx context.Context, req mymodels.UserRegister) (string, bool) {
	return s.api.RegisterUser(ctx, req)
}

func (s *Service) CheckAuth(ctx context.Context, telegramID int64) (string, bool) {
	return s.api.CheckAuth(ctx, telegramID)
}

func (s *Service) ConfirmLogin(ctx context.Context, telegramID int64) (string, bool) {
	return s.api.ConfirmLogin(ctx, telegramID)
}
