package slot

import (
	"app/pkg/models"

	"github.com/google/uuid"
)

func (r *Repository) Create(slot *models.Slot) error {
	if err := r.db.Create(slot).Error; err != nil {
		r.logger.Errorf("Repository.Create (slot): create failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.Create (slot): created id=%d", slot.ID)
	return nil
}

type SlotWithMaster struct {
	models.Slot
	ServiceName      string `json:"service_name"`
	MasterTelegramID int64  `json:"master_telegram_id" gorm:"->;column:master_telegram_id"`
	MasterName       string `json:"master_name" gorm:"->;column:master_name"`
	MasterPhone      string `json:"master_phone" gorm:"->;column:master_phone"`
	MasterSurname    string `json:"master_surname" gorm:"->;column:master_surname"`
	MasterTimezone   string `json:"master_timezone" gorm:"->;column:master_timezone"`
}
type SlotWithMasterAndService struct {
	models.Slot

	ServiceName        string  `json:"service_name" gorm:"->;column:service_name"`
	ServiceDescription string  `json:"service_description" gorm:"->;column:service_description"`
	ServicePrice       float64 `json:"service_price" gorm:"->;column:service_price"`
	ServiceDuration    int     `json:"service_duration" gorm:"->;column:service_duration"`

	MasterTelegramID int64  `json:"master_telegram_id" gorm:"->;column:master_telegram_id"`
	MasterName       string `json:"master_name" gorm:"->;column:master_name"`
	MasterPhone      string `json:"master_phone" gorm:"->;column:master_phone"`
	MasterSurname    string `json:"master_surname" gorm:"->;column:master_surname"`
	MasterTimezone   string `json:"master_timezone" gorm:"->;column:master_timezone"`
}

func (r *Repository) FindSlots(userID uuid.UUID) ([]SlotWithMaster, error) {
	var result []SlotWithMaster

	err := r.db.
		Table("slots").
		Select("slots.*, users.first_name as master_name, users.surname as master_surname, users.phone as master_phone, users.telegram_id as master_telegram_id, users.timezone as master_timezone, services.name as service_name").
		Joins("JOIN users ON users.id = slots.master_id").
		Joins("JOIN services ON services.id = slots.service_id").
		Where("users.id = ?", userID).
		Order("start_time ASC").
		Find(&result).Error
	if err != nil {
		r.logger.Errorf("Repository.FindSlots (slot): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindSlots (slot): master_id=%v count=%d", userID, len(result))
	return result, nil
}

func (r *Repository) FindAllSlots() ([]SlotWithMaster, error) {
	var result []SlotWithMaster
	err := r.db.
		Table("slots").
		Select("slots.*, users.first_name as master_name, users.surname as master_surname, users.phone as master_phone, users.telegram_id as master_telegram_id, users.timezone as master_timezone, services.name as service_name").
		Joins("JOIN users ON users.id = slots.master_id").
		Joins("JOIN services ON services.id = slots.service_id").
		Order("start_time ASC").
		Find(&result).Error
	if err != nil {
		r.logger.Errorf("Repository.FindSlots (slot): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindSlots: count=%d", len(result))
	return result, nil
}
func (r *Repository) FindSlot(slotID uint) (*SlotWithMasterAndService, error) {
	var result *SlotWithMasterAndService
	err := r.db.
		Table("slots").
		Select("slots.*, users.first_name as master_name, users.surname as master_surname, users.phone as master_phone, users.telegram_id as master_telegram_id, users.timezone as master_timezone, services.name as service_name, services.description as service_description, services.price as service_price,services.duration as service_duration").
		Joins("JOIN users ON users.id = slots.master_id").
		Joins("JOIN services ON services.id = slots.service_id").
		Where("slots.id = ?", slotID).
		First(&result).Error
	if err != nil {
		r.logger.Errorf("Repository.FindSlots (slot): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindSlots (slot): slot_id=%v", slotID)
	return result, nil
}
func (r *Repository) DeleteSlots(userID uuid.UUID) error {
	var slots []models.Slot
	if err := r.db.Where("master_id IN (SELECT id FROM users WHERE id = ?)", userID).Delete(&slots).Error; err != nil {
		r.logger.Errorf("Repository.DeleteSlots (slot): delete failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.DeleteSlots (slot): master_id=%v deleted", userID)
	return nil
}

func (r *Repository) DeleteSlot(id uint) error {
	var slot models.Slot
	if err := r.db.Where("id = ?", id).Delete(&slot).Error; err != nil {
		r.logger.Errorf("Repository.DeleteSlot (slot): delete failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.DeleteSlots (slot): master_id=%d deleted", id)
	return nil
}

// GetSlotByIDAndOwner получает слот по ID с проверкой владельца
func (r *Repository) GetSlotByIDAndOwner(slotID uint, ownerID uuid.UUID) (*models.Slot, error) {
	var slot models.Slot
	err := r.db.Where("id = ? AND master_id = ?", slotID, ownerID).First(&slot).Error
	if err != nil {
		r.logger.Errorf("Repository.GetSlotByIDAndOwner (slot): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.GetSlotByIDAndOwner (slot): slot_id=%d owner_id=%v", slotID, ownerID)
	return &slot, nil
}

// CountSlots возвращает общее количество слотов
func (r *Repository) CountSlots() (int64, error) {
	var count int64
	err := r.db.Model(&models.Slot{}).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountSlots: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// CountBookedSlots возвращает количество забронированных слотов
func (r *Repository) CountBookedSlots() (int64, error) {
	var count int64
	err := r.db.Model(&models.Slot{}).Where("is_booked = ?", true).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountBookedSlots: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// GetSlotsByMasterID возвращает слоты мастера
func (r *Repository) GetSlotsByMasterID(masterID uuid.UUID) ([]models.Slot, error) {
	var slots []models.Slot
	err := r.db.Where("master_id = ?", masterID).Find(&slots).Error
	if err != nil {
		r.logger.Errorf("Repository.GetSlotsByMasterID: query failed: %v", err)
		return nil, err
	}
	return slots, nil
}
