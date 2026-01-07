package notification

import (
	"app/pkg/models"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func (s *Service) GetUserNotifications(userID uuid.UUID) ([]models.Notification, error) {
	notifications, err := s.repo.FindUserNotifications(userID)
	if err != nil {
		s.logger.Errorf("Service.GetUserNotifications: repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetUserNotifications: client_id=%s count=%d", userID, len(notifications))
	return notifications, nil
}

func (s *Service) CountUserNotifications(userID uuid.UUID) (int64, error) {
	countNotifications, err := s.repo.CountUnreadUserNotifications(userID)
	if err != nil {
		s.logger.Errorf("Service.CountUserNotifications: repo error: %v", err)
		return 0, err
	}
	s.logger.Infof("Service.CountUserNotifications: client_id=%s count=%d", userID, countNotifications)
	return countNotifications, nil
}

func (s *Service) MarkIsReadNotification(id uint, userID uuid.UUID, isRead bool) error {
	err := s.repo.MarkIsReadNotification(id, userID, isRead)
	if err != nil {
		s.logger.Errorf("Service.MarkIsReadNotification: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.MarkIsReadNotification: id=%d is_read=%t", id, isRead)
	return nil
}

func (s *Service) MarkAllReadNotifications(userID uuid.UUID) error {
	err := s.repo.MarkReadAllNotifications(userID)
	if err != nil {
		s.logger.Errorf("Service.MarkAllReadNotifications: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.MarkAllReadNotifications: user_id=%s", userID)
	return nil
}

// CreateGeneric создает произвольное уведомление для пользователя
func (s *Service) CreateGeneric(userID uuid.UUID, notifType, title, message string, metadata map[string]interface{}) error {
	var meta datatypes.JSON
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			meta = datatypes.JSON(b)
		}
	}
	n := &models.Notification{
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Metadata: meta,
	}
	if err := s.repo.Create(n); err != nil {
		s.logger.Errorf("Service.CreateGeneric: create failed: %v", err)
		return err
	}
	s.logger.Infof("Service.CreateGeneric: user_id=%s type=%s", userID, notifType)
	return nil
}
