package service

import (
	"app/pkg/models"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) CreateService(service *models.Service) error {
	if service.MasterID.String() == "" {
		return fmt.Errorf("MasterID is requiered")
	}
	if err := s.repo.CreateService(service); err != nil {
		s.logger.Errorf("Service.e: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.e: created id=%d", service.ID)
	return nil
}

func (s *Service) GetServices(userID uuid.UUID) ([]models.Service, error) {
	if userID.String() == "" {
		return nil, fmt.Errorf("MasterID is required")
	}
	result, err := s.repo.GetServices(userID)
	if err != nil {
		s.logger.Errorf("Service.GetServices: repo error: %v", err)
		return nil, err
	}
	s.logger.Infof("Service.GetServices: master_id=%v count=%d", userID, len(result))
	return result, nil
}
func (s *Service) GetService(id uint) (models.Service, error) {
	if id == 0 {
		return models.Service{}, fmt.Errorf("MasterID is required")
	}
	result, err := s.repo.GetService(id)
	if err != nil {
		s.logger.Errorf("Service.GetServices: repo error: %v", err)
		return models.Service{}, err
	}
	s.logger.Infof("Service.GetServices: id=%d result=%+v", id, result)
	return result, nil
}
func (s *Service) GetDetailService(id uint) (models.ServiceResponse, error) {
	if id == 0 {
		return models.ServiceResponse{}, fmt.Errorf("MasterID is required")
	}
	result, err := s.repo.GetDetailService(id)
	if err != nil {
		s.logger.Errorf("Service.GetServices: repo error: %v", err)
		return models.ServiceResponse{}, err
	}
	s.logger.Infof("Service.GetServices: id=%d result=%+v", id, result)
	return result, nil
}
func (s *Service) UpdateService(service *models.Service) error {
	if service.ID == 0 {
		return fmt.Errorf("Service ID is required")
	}
	if service.MasterID.String() == "" {
		return fmt.Errorf("MasterID is required")
	}
	if err := s.repo.UpdateService(service); err != nil {
		s.logger.Errorf("Service.UpdateService: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.UpdateService: updated id=%d", service.ID)
	return nil
}

func (s *Service) DeleteService(id uint) error {
	if id == 0 {
		return fmt.Errorf("Service ID is required")
	}
	if err := s.repo.DeleteService(id); err != nil {
		s.logger.Errorf("Service.DeleteService: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.DeleteService: id=%d deleted", id)
	return nil
}

// DeleteServiceByOwner удаляет услугу с проверкой владельца
func (s *Service) DeleteServiceByOwner(serviceID uint, ownerID uuid.UUID) error {
	if serviceID == 0 {
		return fmt.Errorf("Service ID is required")
	}
	if ownerID == uuid.Nil {
		return fmt.Errorf("Owner ID is required")
	}

	// Проверяем, что услуга принадлежит владельцу
	_, err := s.repo.GetServiceByIDAndOwner(serviceID, ownerID)
	if err != nil {
		s.logger.Errorf("Service.DeleteServiceByOwner: service not found or not owned: %v", err)
		return fmt.Errorf("service not found or access denied")
	}

	if err := s.repo.DeleteService(serviceID); err != nil {
		s.logger.Errorf("Service.DeleteServiceByOwner: repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.DeleteServiceByOwner: service_id=%d owner_id=%v deleted", serviceID, ownerID)
	return nil
}

// GetServicesByTelegramID получает услуги по telegram_id
func (s *Service) GetServicesByTelegramID(telegramID int64) ([]models.Service, error) {
	// Находим пользователя по telegram_id
	user, err := s.userRepo.FindByTelegramID(telegramID)
	if err != nil {
		s.logger.Errorf("Service.GetServicesByTelegramID: user lookup failed: %v", err)
		return nil, fmt.Errorf("user not found")
	}
	if user == nil {
		s.logger.Infof("Service.GetServicesByTelegramID: user not found telegram_id=%d", telegramID)
		return nil, fmt.Errorf("user not found")
	}

	// Получаем услуги пользователя
	services, err := s.repo.GetServices(user.ID)
	if err != nil {
		s.logger.Errorf("Service.GetServicesByTelegramID: services lookup failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Service.GetServicesByTelegramID: telegram_id=%d user_id=%v count=%d", telegramID, user.ID, len(services))
	return services, nil
}
