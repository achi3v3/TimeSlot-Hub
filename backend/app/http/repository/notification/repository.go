package notification

import (
	"app/pkg/models"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) Create(notification *models.Notification) error {
	err := r.db.Table("notifications").Create(notification).Error
	if err != nil {
		r.logger.Errorf("Repository.Create (notification): failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.Create (notification): created id=%d type=%s", notification.ID, notification.Type)
	return nil
}
func (r *Repository) FindUserNotifications(userID uuid.UUID) (records []models.Notification, err error) {
	err = r.db.Table("notifications").Where("user_id = ?", userID).
		Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.FindClientNotifications: query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindClientNotifications: user_id=%s count=%d", userID, len(records))
	return
}

func (r *Repository) CountUnreadUserNotifications(userID uuid.UUID) (countNotifications int64, err error) {
	err = r.db.Table("notifications").Where("user_id = ? AND is_read = false", userID).
		Count(&countNotifications).Error
	if err != nil {
		r.logger.Errorf("Repository.CountUserNotifications: query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.CountUserNotifications: user_id=%s count=%d", userID, countNotifications)
	return
}

func (r *Repository) MarkIsReadNotification(id uint, userID uuid.UUID, isRead bool) error {
	result := r.db.Table("notifications").Where("id = ? AND user_id = ?", id, userID).Update("is_read", isRead)
	if result.Error != nil {
		r.logger.Errorf("Repository.MarkIsReadNotification: query failed: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or access denied")
	}
	r.logger.Infof("Repository.MarkIsReadNotification: notification_id=%d, count=%d", id, r.db.RowsAffected)
	return nil
}

func (r *Repository) MarkReadAllNotifications(userID uuid.UUID) (err error) {
	err = r.db.Table("notifications").Where("user_id = ?", userID).Update("is_read", true).Error
	if err != nil {
		r.logger.Errorf("Repository.MarkReadAllNotifications: query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.MarkReadAllNotifications: user_id=%s, count=%d", userID, r.db.RowsAffected)
	return
}
