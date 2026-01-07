package service

import (
	"app/pkg/models"

	"github.com/google/uuid"
)

func (r *Repository) CreateService(slot *models.Service) error {
	if err := r.db.Create(slot).Error; err != nil {
		r.logger.Errorf("Repository.CreateService: create failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.CreateService: created id=%d", slot.ID)
	return nil
}

func (r *Repository) GetServices(userID uuid.UUID) ([]models.Service, error) {
	var services []models.Service
	err := r.db.
		Where("master_id = ?", userID).
		Find(&services).Error
	if err != nil {
		r.logger.Errorf("Repository.GetServices: query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.GetServices: master_id=%v count=%d", userID, len(services))
	return services, nil
}
func (r *Repository) GetAllServices() ([]models.Service, error) {
	var services []models.Service
	err := r.db.
		Find(&services).
		Error
	if err != nil {
		r.logger.Errorf("Repository.GetServices: query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.GetServices: count=%d", len(services))
	return services, nil
}
func (r *Repository) GetService(id uint) (models.Service, error) {
	var service models.Service
	err := r.db.
		Where("id = ?", id).
		First(&service).Error
	if err != nil {
		r.logger.Errorf("Repository.GetService: query failed: %v", err)
		return models.Service{}, err
	}
	r.logger.Infof("Repository.GetService: id=%d result=%+v", id, service)
	return service, nil
}

func (r *Repository) GetDetailService(serviceID uint) (models.ServiceResponse, error) {
	var service models.ServiceResponse
	err := r.db.
		Table("services").
		Select("services.*, users.first_name as master_name, users.surname as master_surname, users.phone as master_phone, users.telegram_id as master_telegram_id").
		Joins("JOIN users ON users.id = services.master_id").
		Where("services.id = ?", serviceID).
		First(&service).Error
	if err != nil {
		r.logger.Errorf("Repository.GetService: query failed: %v", err)
		return models.ServiceResponse{}, err
	}
	return service, nil
}
func (r *Repository) UpdateService(service *models.Service) error {
	if err := r.db.Save(service).Error; err != nil {
		r.logger.Errorf("Repository.UpdateService: update failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.UpdateService: updated id=%d", service.ID)
	return nil
}

func (r *Repository) DeleteService(id uint) error {
	var service models.Service
	if err := r.db.Where("id = ?", id).Delete(&service).Error; err != nil {
		r.logger.Errorf("Repository.DeleteService: delete failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.DeleteService: master_id=%d deleted", id)
	return nil
}

// GetServiceByIDAndOwner получает услугу по ID с проверкой владельца
func (r *Repository) GetServiceByIDAndOwner(serviceID uint, ownerID uuid.UUID) (*models.Service, error) {
	var service models.Service
	err := r.db.Where("id = ? AND master_id = ?", serviceID, ownerID).First(&service).Error
	if err != nil {
		r.logger.Errorf("Repository.GetServiceByIDAndOwner: query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.GetServiceByIDAndOwner: service_id=%d owner_id=%v", serviceID, ownerID)
	return &service, nil
}

// CountServices возвращает общее количество услуг
func (r *Repository) CountServices() (int64, error) {
	var count int64
	err := r.db.Model(&models.Service{}).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountServices: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// GetServicesByMasterID возвращает услуги мастера
func (r *Repository) GetServicesByMasterID(masterID uuid.UUID) ([]models.Service, error) {
	var services []models.Service
	err := r.db.Where("master_id = ?", masterID).Find(&services).Error
	if err != nil {
		r.logger.Errorf("Repository.GetServicesByMasterID: query failed: %v", err)
		return nil, err
	}
	return services, nil
}
