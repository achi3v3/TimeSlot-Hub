package slot

import (
	"app/http/repository/slot"
	"app/http/sender"
	"app/pkg/models"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateSlot(slot *models.Slot) error {
	if slot.MasterID.String() == "" {
		return fmt.Errorf("MasterID is requiered")
	}
	if err := s.repo.Create(slot); err != nil {
		s.logger.Errorf("Service.CreateSlot (slot): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.CreateSlot (slot): created id=%d", slot.ID)
	return nil
}

func (s *Service) GetSlots(userID uuid.UUID) ([]slot.SlotWithMaster, error) {
	if userID.String() == "" {
		return nil, fmt.Errorf("MasterID is required")
	}
	result, err := s.repo.FindSlots(userID)
	if err != nil {
		s.logger.Errorf("Service.GetSlots (slot): repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetSlots (slot): master_id=%v count=%d", userID, len(result))
	return result, nil
}
func (s *Service) GetSlot(slotID uint) (*models.SlotResponse, error) {
	if slotID == 0 {
		return nil, fmt.Errorf("MasterID is required")
	}
	result, err := s.repo.FindSlot(slotID)

	if err != nil {
		s.logger.Errorf("Service.GetSlots (slot): repo error: %v", err)
		return nil, err
	}
	slotResponse := &models.SlotResponse{
		ID:                 result.ID,
		StartTime:          result.StartTime,
		EndTime:            result.EndTime,
		IsBooked:           result.IsBooked,
		ServiceName:        result.ServiceName,
		ServiceDescription: result.ServiceDescription,
		ServiceDuration:    result.ServiceDuration,
		ServicePrice:       result.ServicePrice,
		MasterTelegramID:   result.MasterTelegramID,
		MasterName:         result.MasterName,
		MasterSurname:      result.MasterSurname,
		MasterPhone:        result.MasterPhone,
	}

	s.logger.Infof("Service.GetSlots (slot): slot_id=%v", slotID)
	return slotResponse, nil
}
func (s *Service) DeleteSlots(userID uuid.UUID) error {
	if userID.String() == "" {
		return fmt.Errorf("MasterID is required")
	}
	if err := s.repo.DeleteSlots(userID); err != nil {
		s.logger.Errorf("Service.DeleteSlots (slot): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.DeleteSlots (slot): master_id=%v deleted", userID)
	return nil
}
func (s *Service) DeleteSlot(id uint) error {
	if id == 0 {
		return fmt.Errorf("MasterID is required")
	}
	if err := s.repo.DeleteSlot(id); err != nil {
		s.logger.Errorf("Service.DeleteSlots (slot): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.DeleteSlots (slot): master_id=%d deleted", id)
	return nil
}

// DeleteSlotByOwner удаляет слот с проверкой владельца
func (s *Service) DeleteSlotByOwner(slotID uint, ownerID uuid.UUID) error {
	if slotID == 0 {
		return fmt.Errorf("Slot ID is required")
	}
	if ownerID == uuid.Nil {
		return fmt.Errorf("Owner ID is required")
	}

	// Проверяем, что слот принадлежит владельцу
	_, err := s.repo.GetSlotByIDAndOwner(slotID, ownerID)
	if err != nil {
		s.logger.Errorf("Service.DeleteSlotByOwner (slot): slot not found or not owned: %v", err)
		return fmt.Errorf("slot not found or access denied")
	}

	// ДО удаления: выбираем всех relevant records и детали слота
	var slotDetails *models.Slot
	var confirmRecords, pendingRecords []models.Record
	if s.records != nil {
		sd, _ := s.records.GetSlotByIDWithDetails(slotID)
		slotDetails = &sd
		recs1, _ := s.records.FindRecordsBySlot(slotID, "confirm")
		recs2, _ := s.records.FindRecordsBySlot(slotID, "pending")
		confirmRecords = recs1
		pendingRecords = recs2
	}

	// Потом удаляем слот
	if err := s.repo.DeleteSlot(slotID); err != nil {
		s.logger.Errorf("Service.DeleteSlotByOwner (slot): repo error: %v", err)
		return err
	}

	// Best-effort уведомления клиентам: confirm/pending (site + telegram)
	if s.notify != nil && s.records != nil && slotDetails != nil {
		// подготавливаем локацию для форматирования времени (совместимо с телеграм-сервисом)
		tzName := os.Getenv("TELEGRAM_TIMEZONE")
		if tzName == "" {
			tzName = os.Getenv("TIMEZONE")
		}
		var loc *time.Location
		if tzName != "" {
			if l, err := time.LoadLocation(tzName); err == nil {
				loc = l
			}
		}
		if loc == nil {
			loc = time.Local
		}
		// Общая процедура для всех найденных записей
		notifyForRecords := func(recs []models.Record, st string) {
			for _, r := range recs {
				if r.ClientID == (uuid.UUID{}) {
					continue
				}
				title := "Слот отменен мастером"
				var message string
				if !slotDetails.StartTime.IsZero() && !slotDetails.EndTime.IsZero() {
					message = fmt.Sprintf("Мастер удалил слот %s–%s по услуге \"%s\".\nВаша заявка была в статусе: %s. Свяжитесь с мастером при необходимости.",
						slotDetails.StartTime.In(loc).Format("02.01.2006 15:04"),
						slotDetails.EndTime.In(loc).Format("15:04"),
						slotDetails.Service.Name,
						st,
					)
				} else {
					message = fmt.Sprintf("Мастер удалил слот, на который у вас была заявка. Статус заявки: %s.", st)
				}
				meta := map[string]interface{}{
					"record_id": r.ID,
					"slot_id":   r.SlotID,
					"status":    r.Status,
				}
				_ = s.notify.CreateGeneric(r.ClientID, "SLOT_DELETED", title, message, meta)
				client, err := s.records.GetUserByID(r.ClientID)
				if err == nil && client.TelegramID != 0 {
					_ = sender.RecordStatusNotify(client.TelegramID, title, message)
				}
			}
		}
		notifyForRecords(confirmRecords, "confirm")
		notifyForRecords(pendingRecords, "pending")
	}
	return nil
}
