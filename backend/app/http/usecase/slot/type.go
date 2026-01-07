package slot

import (
	recordRepo "app/http/repository/record"
	"app/http/repository/slot"
	notifyServ "app/http/usecase/notification"

	"github.com/sirupsen/logrus"
)

type Service struct {
	repo    *slot.Repository
	logger  *logrus.Logger
	notify  *notifyServ.Service
	records *recordRepo.Repository
}

func NewService(repo *slot.Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// WithNotification подключает сервис уведомлений к сервису слотов
func (s *Service) WithNotification(ns *notifyServ.Service) *Service {
	s.notify = ns
	return s
}

// WithRecordRepository подключает репозиторий заявок
func (s *Service) WithRecordRepository(r *recordRepo.Repository) *Service {
	s.records = r
	return s
}
