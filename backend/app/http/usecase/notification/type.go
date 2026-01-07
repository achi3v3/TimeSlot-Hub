package notification

import (
	"app/http/repository/notification"
	"app/pkg/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo    *notification.Repository
	factory *NotificationFactory
	logger  *logrus.Logger
}

func NewService(repo *notification.Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:    repo,
		factory: &NotificationFactory{},
		logger:  logger,
	}
}

func (s *Service) CreateRecordCreatedNotification(masterID uuid.UUID, record *models.Record, clientName, clientSurname string, slot *models.Slot, service *models.Service, master *models.User) error {
	notification := s.factory.CreateRecordCreated(masterID, record, clientName, clientSurname, slot, service, master)
	return s.repo.Create(notification)
}

func (s *Service) CreateRecordStatusNotification(clientID uuid.UUID, record *models.Record, status string, slot *models.Slot, service *models.Service, master *models.User) error {
	notification := s.factory.CreateRecordStatus(clientID, record, status, slot, service, master)
	return s.repo.Create(notification)
}
