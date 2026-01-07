package slots

import (
	"context"
	"fmt"
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

func (s *Service) GetByMasterID(ctx context.Context, masterID int64) ([]mymodels.SlotResponse, error) {
	slots, ok := s.api.GetSlotsByTelegramID(ctx, masterID)
	if !ok {
		return nil, fmt.Errorf("failed to fetch slots")
	}
	return slots, nil
}

func (s *Service) DeleteByMasterID(ctx context.Context, masterID uint) error {
	if ok := s.api.DeleteSlotsByTelegramID(ctx, masterID); !ok {
		return fmt.Errorf("failed to delete slots")
	}
	return nil
}
