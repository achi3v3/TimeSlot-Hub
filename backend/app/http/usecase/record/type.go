package record

import (
	"app/http/repository/record"
	"app/http/usecase/notification"

	"github.com/sirupsen/logrus"
)

type Service struct {
	repo                *record.Repository
	notificationService *notification.Service
	logger              *logrus.Logger
}

func NewService(repo *record.Repository, notificationService *notification.Service, logger *logrus.Logger) *Service {
	return &Service{
		repo:                repo,
		notificationService: notificationService,
		logger:              logger,
	}
}
