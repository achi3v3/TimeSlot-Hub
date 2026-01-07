package record

import (
	"app/http/sender"
	"app/pkg/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Service) Create(book *models.Record) error {
	// Проверяем, не существует ли уже запись от этого пользователя на этот слот
	exists, err := s.repo.ExistsRecord(book.SlotID, book.ClientID)
	if err != nil {
		s.logger.Errorf("Service.Create (record): check existing failed: %v", err)
		return err
	}
	if exists {
		s.logger.Errorf("Service.Create (record): record already exists for slot_id=%d client_id=%s", book.SlotID, book.ClientID)
		return fmt.Errorf("user already has a record for this slot")
	}

	bookID, err := s.repo.Create(book)
	if err != nil {
		s.logger.Errorf("Service.Create (record): repo error: %v", err)
		return err
	}

	// try send notification to master about new record (site notification)
	// Load slot with details to get master, service info
	slot, err := s.repo.GetSlotByIDWithDetails(book.SlotID)
	if err != nil {
		s.logger.Errorf("Service.Create (record): load slot failed for notification: %v", err)
	} else {
		// Load client for name info
		client, err := s.repo.GetUserByID(book.ClientID)
		if err != nil {
			s.logger.Errorf("Service.Create (record): load client failed for notification: %v", err)
		} else {
			if err := s.notificationService.CreateRecordCreatedNotification(slot.MasterID, book, client.FirstName, client.Surname, &slot, &slot.Service, &slot.Master); err != nil {
				s.logger.Errorf("Service.Create (record): send notification failed: %v", err)
			}
			// telegram notify to master (best-effort) with detailed message
			if slot.Master.TelegramID != 0 {
				title := "Новая запись от клиента"
				// Время в таймзоне мастера
				loc := time.Local
				if slot.Master.Timezone != "" {
					if l, err := time.LoadLocation(slot.Master.Timezone); err == nil {
						loc = l
					}
				}
				start := slot.StartTime.In(loc).Format("02.01.2006 15:04")
				end := slot.EndTime.In(loc).Format("15:04")
				tzLabel := slot.Master.Timezone
				// Добавляем телефон клиента для верификации личности
				message := fmt.Sprintf("Клиент %s %s (тел: %s) записался на услугу \"%s\" (%s руб.)\nВремя: %s - %s (%s)",
					client.FirstName, client.Surname, client.Phone, slot.Service.Name, fmt.Sprintf("%.0f", slot.Service.Price),
					start, end, tzLabel)
				_ = sender.RecordNotify(bookID, slot.Master.TelegramID, title, message)
			}
		}
	}

	s.logger.Infof("Service.Create (record): created id=%d", book.ID)
	return nil
}

func (s *Service) GetClientRecords(client_id uuid.UUID) ([]models.Record, error) {
	records, err := s.repo.FindRecordsByClient(client_id)
	if err != nil {
		s.logger.Errorf("Service.GetBooks (record): repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetBooks (record): client_id=%s count=%d", client_id, len(records))
	return records, nil
}

// GetClientRecordsByStatus returns client records filtered by status (optional)
func (s *Service) GetClientRecordsByStatus(clientID uuid.UUID, status string) ([]models.Record, error) {
	records, err := s.repo.FindRecordsByClientWithStatus(clientID, status)
	if err != nil {
		s.logger.Errorf("Service.GetClientRecordsByStatus: repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetClientRecordsByStatus: client_id=%s status=%s count=%d", clientID, status, len(records))
	return records, nil
}

func (s *Service) GetRecordsBySlot(slot_id uint, status string) ([]models.Record, error) {
	records, err := s.repo.FindRecordsBySlot(slot_id, status)
	if err != nil {
		s.logger.Errorf("Service.GetRecordsBySlot: repo error: %v", err)
		return records, err
	}
	s.logger.Infof("Service.GetRecordsBySlot: slot_id=%d", slot_id)
	return records, nil
}

func (s *Service) GetDetailRecord(recordID uint) (models.RecordResponce, error) {
	record, err := s.repo.FindDetailRecord(recordID)
	if err != nil {
		s.logger.Errorf("Service.GetRecordsBySlot: repo error: %v", err)
		return record, err
	}
	s.logger.Infof("Service.GetRecordsBySlot: record=%+v", record)
	return record, nil
}
func (s *Service) GetAllRecordsBySlot(slot_id uint) ([]models.Record, error) {
	records, err := s.repo.FindAllRecordsBySlot(slot_id)
	if err != nil {
		s.logger.Errorf("Service.GetAllRecordsBySlot: repo error: %v", err)
		return records, err
	}
	s.logger.Infof("Service.GetAllRecordsBySlot: slot_id=%d", slot_id)
	return records, nil
}
func (s *Service) ConfirmRecord(record_id uint) error {
	// confirm in repository
	if err := s.repo.ChangeRecordStatus(record_id, "confirm"); err != nil {
		s.logger.Errorf("Service.ConfirmRecord: repo error: %v", err)
		return err
	}
	// load record with details to notify client (site)
	rec, err := s.repo.GetRecordByIDWithDetails(record_id)
	if err != nil {
		s.logger.Errorf("Service.ConfirmRecord: load record failed for notification: %v", err)
	} else {
		if err := s.notificationService.CreateRecordStatusNotification(rec.ClientID, &rec, "confirm", &rec.Slot, &rec.Slot.Service, &rec.Slot.Master); err != nil {
			s.logger.Errorf("Service.ConfirmRecord: send notification failed: %v", err)
		}
		// telegram notify to client (best-effort) with detailed message
		client, err2 := s.repo.GetUserByID(rec.ClientID)
		if err2 == nil && client.TelegramID != 0 {
			title := "Запись подтверждена ✅"
			// Время в таймзоне мастера
			loc := time.Local
			if rec.Slot.Master.Timezone != "" {
				if l, err := time.LoadLocation(rec.Slot.Master.Timezone); err == nil {
					loc = l
				}
			}
			start := rec.Slot.StartTime.In(loc).Format("02.01.2006 15:04")
			end := rec.Slot.EndTime.In(loc).Format("15:04")
			message := fmt.Sprintf("Мастер подтвердил вашу запись\n\nУслуга: %s (%s руб.)\nМастер: %s %s\nВремя: %s - %s (%s)",
				rec.Slot.Service.Name, fmt.Sprintf("%.0f", rec.Slot.Service.Price),
				rec.Slot.Master.FirstName, rec.Slot.Master.Surname,
				start, end, rec.Slot.Master.Timezone)
			_ = sender.RecordStatusNotify(client.TelegramID, title, message)
		}
	}
	s.logger.Infof("Service.ConfirmRecord: record_id=%d confirmed", record_id)
	return nil
}
func (s *Service) RejectRecord(record_id uint) error {
	// confirm in repository
	if err := s.repo.ChangeRecordStatus(record_id, "reject"); err != nil {
		s.logger.Errorf("Service.RejectRecord: repo error: %v", err)
		return err
	}
	// load record with details to notify client (site)
	rec, err := s.repo.GetRecordByIDWithDetails(record_id)
	if err != nil {
		s.logger.Errorf("Service.RejectRecord: load record failed for notification: %v", err)
	} else {
		if err := s.notificationService.CreateRecordStatusNotification(rec.ClientID, &rec, "confirm", &rec.Slot, &rec.Slot.Service, &rec.Slot.Master); err != nil {
			s.logger.Errorf("Service.RejectRecord: send notification failed: %v", err)
		}
		// telegram notify to client (best-effort) with detailed message
		client, err2 := s.repo.GetUserByID(rec.ClientID)
		if err2 == nil && client.TelegramID != 0 {
			title := "Запись отклонена ❌"
			loc := time.Local
			if rec.Slot.Master.Timezone != "" {
				if l, err := time.LoadLocation(rec.Slot.Master.Timezone); err == nil {
					loc = l
				}
			}
			start := rec.Slot.StartTime.In(loc).Format("02.01.2006 15:04")
			end := rec.Slot.EndTime.In(loc).Format("15:04")
			message := fmt.Sprintf("Мастер отклонил вашу запись\n\nУслуга: %s (%s руб.)\nМастер: %s %s\nВремя: %s - %s (%s)",
				rec.Slot.Service.Name, fmt.Sprintf("%.0f", rec.Slot.Service.Price),
				rec.Slot.Master.FirstName, rec.Slot.Master.Surname,
				start, end, rec.Slot.Master.Timezone)
			_ = sender.RecordStatusNotify(client.TelegramID, title, message)
		}
	}
	s.logger.Infof("Service.RejectRecord: record_id=%d confirmed", record_id)
	return nil
}
func (s *Service) DeleteRecord(id uint) error {
	if err := s.repo.DeleteRecord(id); err != nil {
		s.logger.Errorf("Service.DeleteBook (record): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.DeleteBook (record): deleted id=%d", id)
	return nil
}

func (s *Service) UpdateRecordStatus(recordID uint, status string) error {
	if status != "pending" && status != "confirm" && status != "reject" {
		s.logger.Errorf("Service.UpdateRecordStatus: invalid status: %s", status)
		return fmt.Errorf("invalid status")
	}
	if err := s.repo.UpdateRecordStatus(recordID, status); err != nil {
		s.logger.Errorf("Service.UpdateRecordStatus: repo error: %v", err)
		return err
	}
	// load record with details to notify client (site)
	rec, err := s.repo.GetRecordByIDWithDetails(recordID)
	if err != nil {
		s.logger.Errorf("Service.UpdateRecordStatus: load record failed for notification: %v", err)
	} else if status == "confirm" || status == "reject" {
		if err := s.notificationService.CreateRecordStatusNotification(rec.ClientID, &rec, status, &rec.Slot, &rec.Slot.Service, &rec.Slot.Master); err != nil {
			s.logger.Errorf("Service.UpdateRecordStatus: send notification failed: %v", err)
		}
		// telegram notify to client (best-effort) with detailed message
		client, err2 := s.repo.GetUserByID(rec.ClientID)
		if err2 == nil && client.TelegramID != 0 {
			var title, emoji string
			if status == "confirm" {
				title = "Запись подтверждена ✅"
				emoji = "подтвердил"
			} else {
				title = "Запись отклонена ❌"
				emoji = "отклонил"
			}
			// Форматируем в таймзоне мастера
			loc := time.Local
			if rec.Slot.Master.Timezone != "" {
				if l, err := time.LoadLocation(rec.Slot.Master.Timezone); err == nil {
					loc = l
				}
			}
			start := rec.Slot.StartTime.In(loc).Format("02.01.2006 15:04")
			end := rec.Slot.EndTime.In(loc).Format("15:04")
			message := fmt.Sprintf("Мастер %s вашу запись\n\nУслуга: %s (%s руб.)\nМастер: %s %s\nВремя: %s - %s (%s)",
				emoji, rec.Slot.Service.Name, fmt.Sprintf("%.0f", rec.Slot.Service.Price),
				rec.Slot.Master.FirstName, rec.Slot.Master.Surname,
				start, end, rec.Slot.Master.Timezone)
			_ = sender.RecordStatusNotify(client.TelegramID, title, message)
		}
	}
	s.logger.Infof("Service.UpdateRecordStatus: record_id=%d status=%s", recordID, status)
	return nil
}

// GetUpcomingRecordsByMasterTelegramID возвращает предстоящие подтвержденные записи мастера
func (s *Service) GetUpcomingRecordsByMasterTelegramID(masterTelegramID int64) ([]models.Record, error) {
	records, err := s.repo.FindUpcomingRecordsByMasterTelegramID(masterTelegramID)
	if err != nil {
		s.logger.Errorf("Service.GetUpcomingRecordsByMasterTelegramID: repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetUpcomingRecordsByMasterTelegramID: master_telegram_id=%d count=%d", masterTelegramID, len(records))
	return records, nil
}
