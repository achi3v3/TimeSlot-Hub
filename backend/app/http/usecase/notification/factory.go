package notification

import (
	"app/pkg/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationFactory struct{}

func (f *NotificationFactory) CreateRecordCreated(masterID uuid.UUID, record *models.Record, firstName, surname string, slot *models.Slot, service *models.Service, master *models.User) *models.Notification {
	metaData := map[string]interface{}{
		"record_id":     record.ID,
		"slot_id":       record.SlotID,
		"client_id":     record.ClientID,
		"client_name":   fmt.Sprintf("%s %s", firstName, surname),
		"master_id":     masterID,
		"master_name":   fmt.Sprintf("%s %s", master.FirstName, master.Surname),
		"service_id":    service.ID,
		"service_name":  service.Name,
		"service_price": service.Price,
		"slot_start":    slot.StartTime.Format("02.01.2006 15:04"),
		"slot_end":      slot.EndTime.Format("02.01.2006 15:04"),
		"action_url":    fmt.Sprintf("records/%d", record.ID),
	}

	title := "Новая запись от клиента"
	message := fmt.Sprintf("Клиент %s %s записался на услугу \"%s\" (%s руб.)\nВремя: %s - %s",
		firstName, surname, service.Name, fmt.Sprintf("%.0f", service.Price),
		slot.StartTime.Format("02.01.2006 15:04"), slot.EndTime.Format("15:04"))

	return &models.Notification{
		UserID:    masterID,
		Type:      "RECORD_CREATED",
		Title:     title,
		Message:   message,
		Metadata:  f.toJSON(metaData),
		ExpiresAt: f.expiresIn(30 * 24 * time.Hour),
	}
}

func (f *NotificationFactory) CreateRecordStatus(clientID uuid.UUID, record *models.Record, status string, slot *models.Slot, service *models.Service, master *models.User) *models.Notification {
	configs := map[string]struct {
		title     string
		message   string
		notifType string
	}{
		"confirm": {"Запись подтверждена ✅", "Мастер подтвердил вашу запись", "RECORD_CONFIRMED"},
		"reject":  {"Запись отклонена ❌", "Мастер отклонил вашу запись", "RECORD_REJECTED"},
	}
	config := configs[status]

	metadata := map[string]interface{}{
		"record_id":     record.ID,
		"slot_id":       record.SlotID,
		"status":        status,
		"master_id":     master.ID,
		"master_name":   fmt.Sprintf("%s %s", master.FirstName, master.Surname),
		"service_id":    service.ID,
		"service_name":  service.Name,
		"service_price": service.Price,
		"slot_start":    slot.StartTime.Format("02.01.2006 15:04"),
		"slot_end":      slot.EndTime.Format("02.01.2006 15:04"),
		"action_url":    fmt.Sprintf("/my-records/%d", record.ID),
	}

	// Создаем подробное сообщение
	message := fmt.Sprintf("%s\n\nУслуга: %s (%s руб.)\nМастер: %s %s\nВремя: %s - %s",
		config.message, service.Name, fmt.Sprintf("%.0f", service.Price),
		master.FirstName, master.Surname,
		slot.StartTime.Format("02.01.2006 15:04"), slot.EndTime.Format("15:04"))

	return &models.Notification{
		UserID:    clientID,
		Type:      config.notifType,
		Title:     config.title,
		Message:   message,
		Metadata:  f.toJSON(metadata),
		ExpiresAt: f.expiresIn(15 * 24 * time.Hour),
	}
}

func (f *NotificationFactory) toJSON(data map[string]interface{}) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return datatypes.JSON(jsonData)
}

func (f *NotificationFactory) expiresIn(hours time.Duration) *time.Time {
	t := time.Now().Add(hours)
	return &t
}
