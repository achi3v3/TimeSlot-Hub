package record

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

func (s *Service) GetByMasterID(ctx context.Context, masterID int64) ([]mymodels.Record, error) {
	records, err := s.api.GetRecords(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slots")
	}
	return records, nil
}
