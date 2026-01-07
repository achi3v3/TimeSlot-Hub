package record

import (
	"app/pkg/models"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

func (r *Repository) Create(book *models.Record) (uint, error) {
	exists, err := r.ExistsRecord(book.SlotID, book.ClientID)
	if err != nil {
		r.logger.Errorf("Repository.Create (record): check existing failed: %v", err)
		return 0, err
	}
	if exists {
		r.logger.Errorf("Repository.Create (record): record already exists for slot_id=%d client_id=%s", book.SlotID, book.ClientID)
		return 0, fmt.Errorf("user already has a record for this slot")
	}

	if err := r.db.Create(book).Error; err != nil {
		r.logger.Errorf("Repository.Create (record): create failed: %v", err)
		return 0, err
	}
	r.logger.Infof("Repository.Create (record): created id=%d", book.ID)
	return book.ID, nil
}

func (r *Repository) FindRecordsByClient(client_id uuid.UUID) (records []models.Record, err error) {
	err = r.db.Preload("Slot.Service").Preload("Slot.Master").
		Where("client_id = ?", client_id).
		Order("id DESC").
		Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.FindBooksByClient (record): query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindBooksByClient (record): client_id=%s count=%d", client_id, len(records))
	return
}

// FindRecordsByClientWithStatus returns records for a client optionally filtered by status
func (r *Repository) FindRecordsByClientWithStatus(clientID uuid.UUID, status string) (records []models.Record, err error) {
	q := r.db.Preload("Slot.Service").Preload("Slot.Master")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err = q.Where("client_id = ?", clientID).
		Order("id DESC").
		Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.FindRecordsByClientWithStatus: query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindRecordsByClientWithStatus: client_id=%s status=%s count=%d", clientID, status, len(records))
	return
}

func (r *Repository) FindRecordsBySlot(slot_id uint, status string) (records []models.Record, err error) {
	err = r.db.Preload("Client").Where("slot_id = ? AND status = ?", slot_id, status).Order("id ASC").Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.FindRecordBySlot (record): query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindRecordBySlot (record): slot_id=%d", slot_id)
	return
}
func (r *Repository) FindDetailRecord(recordID uint) (record models.RecordResponce, err error) {
	err = r.db.Table("records").
		Select(`
			records.id,
			records.status,
			records.created_at,
			records.client_id,
			clients.telegram_id as client_telegram_id,
			clients.first_name as client_name,
			clients.surname as client_surname,
			clients.phone as client_phone,
			records.slot_id,
			services.name as slot_name,
			services.price as slot_price,
			services.duration as slot_duration,
			slots.master_id,
			masters.telegram_id as master_telegram_id,
			masters.first_name as master_name,
			masters.surname as master_surname,
			masters.phone as master_phone
		`).
		Joins("LEFT JOIN users as clients ON records.client_id = clients.id").
		Joins("LEFT JOIN slots ON records.slot_id = slots.id").
		Joins("LEFT JOIN services ON slots.service_id = services.id").
		Joins("LEFT JOIN users as masters ON slots.master_id = masters.id").
		Where("records.id = ?", recordID).
		Scan(&record).Error
	if err != nil {
		r.logger.Errorf("Repository.FindDetailRecord: query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindDetailRecord: record_id=%d", recordID)
	return
}
func (r *Repository) FindAllRecordsBySlot(slot_id uint) (records []models.Record, err error) {
	err = r.db.Preload("Client").Where("slot_id = ?", slot_id).Order("id ASC").Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.FindRecordBySlot (record): query failed: %v", err)
		return
	}
	r.logger.Infof("Repository.FindRecordBySlot (record): slot_id=%d", slot_id)
	return
}
func (r *Repository) ChangeRecordStatus(record_id uint, status string) (err error) {
	// Транзакция: меняем статус записи, если подтвердили отклоняем остальные по этому слоту, помечаем слот занятым
	return r.db.Transaction(func(tx *gorm.DB) error {
		var rec models.Record
		if err := tx.First(&rec, record_id).Error; err != nil {
			r.logger.Errorf("Repository.ChangeRecordStatus: load record failed: %v", err)
			return err
		}

		if err := tx.Model(&models.Record{}).Where("id = ?", record_id).Update("status", status).Error; err != nil {
			r.logger.Errorf("Repository.ChangeRecordStatus: set confirm failed: %v", err)
			return err
		}

		// Отклоняем все остальные записи по этому слоту
		if status == "confirm" {
			if err := tx.Model(&models.Record{}).
				Where("slot_id = ? AND id <> ? AND status <> ?", rec.SlotID, record_id, "reject").
				Update("status", "reject").Error; err != nil {
				r.logger.Errorf("Repository.ChangeRecordStatus: reject others failed: %v", err)
				return err
			}

			// Помечаем слот занятым
			if err := tx.Model(&models.Slot{}).Where("id = ?", rec.SlotID).Update("is_booked", true).Error; err != nil {
				r.logger.Errorf("Repository.ChangeRecordStatus: mark slot booked failed: %v", err)
				return err
			}
		}

		r.logger.Infof("Repository.ChangeRecordStatus: record_id=%d %s, slot_id=%d booked", record_id, status, rec.SlotID)
		return nil
	})
}

func (r *Repository) DeleteRecord(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var rec models.Record
		if err := tx.First(&rec, id).Error; err != nil {
			r.logger.Errorf("Repository.DeleteBook: load record failed: %v", err)
			return err
		}

		if err := tx.Where("id = ?", id).Delete(&models.Record{}).Error; err != nil {
			r.logger.Errorf("Repository.DeleteBook: delete failed: %v", err)
			return err
		}

		if rec.Status == "confirm" {
			var cnt int64
			if err := tx.Model(&models.Record{}).Where("slot_id = ? AND status = ?", rec.SlotID, "confirm").Count(&cnt).Error; err != nil {
				return err
			}
			booked := cnt > 0
			if err := tx.Model(&models.Slot{}).Where("id = ?", rec.SlotID).Update("is_booked", booked).Error; err != nil {
				return err
			}
		}

		r.logger.Infof("Repository.DeleteBook: deleted id=%d", id)
		return nil
	})
}

// GetSlotByIDWithDetails returns slot by id with service and master loaded
func (r *Repository) GetSlotByIDWithDetails(id uint) (models.Slot, error) {
	var slot models.Slot
	if err := r.db.Preload("Service").Preload("Master").First(&slot, id).Error; err != nil {
		r.logger.Errorf("Repository.GetSlotByIDWithDetails: load failed: %v", err)
		return slot, err
	}
	return slot, nil
}

// GetRecordByIDWithDetails returns a record by id with slot, service, master and client loaded
func (r *Repository) GetRecordByIDWithDetails(id uint) (models.Record, error) {
	var rec models.Record
	if err := r.db.Preload("Slot.Service").Preload("Slot.Master").Preload("Client").First(&rec, id).Error; err != nil {
		r.logger.Errorf("Repository.GetRecordByIDWithDetails: load failed: %v", err)
		return rec, err
	}
	return rec, nil
}

// GetSlotByID returns slot by id
func (r *Repository) GetSlotByID(id uint) (models.Slot, error) {
	var slot models.Slot
	if err := r.db.First(&slot, id).Error; err != nil {
		r.logger.Errorf("Repository.GetSlotByID: load failed: %v", err)
		return slot, err
	}
	return slot, nil
}

// GetUserByID returns user by id
func (r *Repository) GetUserByID(id uuid.UUID) (models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		r.logger.Errorf("Repository.GetUserByID: load failed: %v", err)
		return user, err
	}
	return user, nil
}

// UpdateRecordStatus устанавливает статус записи с учетом инвариантов:
// - confirm: отклоняет остальные записи слота и помечает слот занятым
// - reject/pending: обновляет статус записи; если запись была confirm и стала не confirm, то пересчитывает is_booked
func (r *Repository) UpdateRecordStatus(recordID uint, status string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var rec models.Record
		if err := tx.First(&rec, recordID).Error; err != nil {
			r.logger.Errorf("Repository.UpdateRecordStatus: load record failed: %v", err)
			return err
		}

		prevStatus := rec.Status
		if err := tx.Model(&models.Record{}).Where("id = ?", recordID).Update("status", status).Error; err != nil {
			r.logger.Errorf("Repository.UpdateRecordStatus: update status failed: %v", err)
			return err
		}

		if status == "confirm" {
			if err := tx.Model(&models.Record{}).
				Where("slot_id = ? AND id <> ? AND status <> ?", rec.SlotID, recordID, "reject").
				Update("status", "reject").Error; err != nil {
				r.logger.Errorf("Repository.UpdateRecordStatus: reject others failed: %v", err)
				return err
			}
			if err := tx.Model(&models.Slot{}).Where("id = ?", rec.SlotID).Update("is_booked", true).Error; err != nil {
				r.logger.Errorf("Repository.UpdateRecordStatus: mark slot booked failed: %v", err)
				return err
			}
		} else {
			// Если запись была подтверждена и стала НЕ подтверждена, возможно слоту нужно снять флаг is_booked
			if prevStatus == "confirm" {
				var cnt int64
				if err := tx.Model(&models.Record{}).Where("slot_id = ? AND status = ?", rec.SlotID, "confirm").Count(&cnt).Error; err != nil {
					return err
				}
				booked := cnt > 0
				if err := tx.Model(&models.Slot{}).Where("id = ?", rec.SlotID).Update("is_booked", booked).Error; err != nil {
					return err
				}
			}
		}

		r.logger.Infof("Repository.UpdateRecordStatus: record_id=%d status=%s", recordID, status)
		return nil
	})
}

// ExistsRecord проверяет, существует ли уже запись от пользователя на слот
func (r *Repository) ExistsRecord(slotID uint, clientID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Record{}).Where("slot_id = ? AND client_id = ?", slotID, clientID).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.ExistsRecord: query failed: %v", err)
		return false, err
	}
	return count > 0, nil
}

// CountRecords возвращает общее количество записей
func (r *Repository) CountRecords() (int64, error) {
	var count int64
	err := r.db.Model(&models.Record{}).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountRecords: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// CountRecordsByStatus возвращает количество записей по статусу
func (r *Repository) CountRecordsByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Record{}).Where("status = ?", status).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountRecordsByStatus: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// GetRecordsByMasterID возвращает записи мастера
func (r *Repository) GetRecordsByMasterID(masterID uuid.UUID) ([]models.Record, error) {
	var records []models.Record
	err := r.db.
		Joins("JOIN slots ON slots.id = records.slot_id").
		Where("slots.master_id = ?", masterID).
		Order("records.created_at DESC").
		Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.GetRecordsByMasterID: query failed: %v", err)
		return nil, err
	}
	return records, nil
}
func (r *Repository) GetAllRecords() ([]models.Record, error) {
	var records []models.Record
	err := r.db.
		Order("records.created_at DESC").
		Find(&records).Error
	if err != nil {
		r.logger.Errorf("Repository.GetRecordsByMasterID: query failed: %v", err)
		return nil, err
	}
	return records, nil
}

// FindConfirmedRecordsStartingBetween finds confirmed records whose slot starts within [from, to].
// Preloads Slot.Service, Slot.Master, and Client for composing reminders.
func (r *Repository) FindConfirmedRecordsStartingBetween(from, to time.Time) ([]models.Record, error) {
	var records []models.Record
	q := r.db.
		Preload("Slot.Service").
		Preload("Slot.Master").
		Preload("Client").
		Joins("JOIN slots ON slots.id = records.slot_id").
		Where("records.status = ? AND slots.start_time BETWEEN ? AND ?", "confirm", from, to)
	if err := q.Find(&records).Error; err != nil {
		r.logger.Errorf("Repository.FindConfirmedRecordsStartingBetween: query failed: %v", err)
		return nil, err
	}
	return records, nil
}

// FindUpcomingRecordsByMasterTelegramID возвращает предстоящие записи для мастера по его telegram_id
func (r *Repository) FindUpcomingRecordsByMasterTelegramID(masterTelegramID int64) ([]models.Record, error) {
	var records []models.Record
	now := time.Now()

	err := r.db.
		Preload("Slot.Service").
		Preload("Slot.Master").
		Preload("Client").
		Joins("JOIN slots ON slots.id = records.slot_id").
		Joins("JOIN users ON slots.master_id = users.id").
		Where("users.telegram_id = ? AND slots.start_time > ? AND records.status = ?", masterTelegramID, now, "confirm").
		Order("slots.start_time ASC").
		Find(&records).Error

	if err != nil {
		r.logger.Errorf("Repository.FindUpcomingRecordsByMasterTelegramID: query failed: %v", err)
		return nil, err
	}

	r.logger.Infof("Repository.FindUpcomingRecordsByMasterTelegramID: master_telegram_id=%d count=%d", masterTelegramID, len(records))
	return records, nil
}
