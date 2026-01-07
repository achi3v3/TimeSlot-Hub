package user

import (
	"app/encoder"
	"app/http/sender"
	"app/pkg/models"
	"fmt"

	"github.com/google/uuid"
)

// Register отправляет запрос в бд на создание пользователя.
// Ответ: Возвращает ошибку.
func (s *Service) Register(user *models.User) error {
	if user.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	user.Roles = nil
	if err := s.repo.Create(user); err != nil {
		s.logger.Errorf("Service.Register (user): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.Register (user): created id=%s", user.ID)
	return nil
}

// Login осуществляет вход пользователя по номеру телефона.
// Ответ: Возвращает структуру User, сообщение об операции и ошибку.
func (s *Service) Login(phone string) (*models.User, string, error) {
	// Валидация данных
	var message string
	if phone == "" {
		return nil, message, fmt.Errorf("phone is required")
	}

	user, err := s.repo.FindByPhone(phone)
	if err != nil {
		s.logger.Errorf("Service.Login (user): repo error: %v", err)
		return nil, message, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		s.logger.Infof("Service.Login (user): user not found phone=%s", phone)
		return nil, "User not found", fmt.Errorf("user not found")
	}

	// Уведомление отправляет контроллер, чтобы включить IP/локацию
	return user, "Confirmation required. Please check Telegram", nil
}

func (s *Service) GetByTelegramID(telegram_id int64) (*models.User, error) {
	user, err := s.repo.FindByTelegramID(telegram_id)
	if err != nil {
		s.logger.Errorf("Service.GetByTelegramID (user): repo error: %v", err)
		return nil, err
	}
	if user == nil {
		s.logger.Infof("Service.GetByTelegramID (user): not found telegram_id=%d", telegram_id)
		return nil, nil
	}
	s.logger.Infof("Service.GetByTelegramID (user): found id=%s", user.ID)
	return user, nil
}

func (s *Service) ConfirmLoginByTelegramID(telegram_id int64) error {
	user, err := s.repo.FindByTelegramID(telegram_id)
	if err != nil {
		s.logger.Errorf("Service.GetByTelegramID (user): repo error: %v", err)
		return err
	}
	token, err := encoder.GenerateToken(user.ID)
	if err != nil {
		s.logger.Errorf("Service.ConfirmLoginByTelegramID: token generation failed: %v", err)
		return err
	}
	if err := s.repo.StorageToken(telegram_id, token); err != nil {
		s.logger.Errorf("Service.ConfirmLoginByTelegramID: storage token failed: %v", err)
		return err
	}
	s.logger.Infof("Service.ConfirmLoginByTelegramID (user): found id=%s", user.ID)
	return nil
}

func (s *Service) ClaimUserTokenByTelegramID(telegram_id int64) (string, error) {
	token, err := s.repo.ClaimUserToken(telegram_id)
	if err != nil {
		s.logger.Errorf("Service.ConfirmLoginByTelegramID: storage token failed: %v", err)
		return "", err
	}
	s.logger.Infof("Service.ConfirmLoginByTelegramID (user): found token for user with telegram_id=%d", telegram_id)
	return token, nil
}

func (s *Service) CheckUserTokenByTelegramID(telegram_id int64) bool {
	ok := s.repo.CheckUserToken(telegram_id)
	return ok
}

// UpdateNames обновляет только first_name и surname пользователя
func (s *Service) UpdateNames(req UpdateNamesRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	id, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	if req.FirstName == "" && req.Surname == "" {
		return fmt.Errorf("nothing to update")
	}
	if err := s.repo.UpdateNames(id, req.FirstName, req.Surname); err != nil {
		s.logger.Errorf("Service.UpdateNames (user): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.UpdateNames (user): updated id=%s", id)
	return nil
}

// UpdateTimezone обновляет таймзону пользователя
func (s *Service) UpdateTimezone(req UpdateTimezoneRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}
	id, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	if err := s.repo.UpdateTimezone(id, req.Timezone); err != nil {
		s.logger.Errorf("Service.UpdateTimezone (user): repo error: %v", err)
		return err
	}
	s.logger.Infof("Service.UpdateTimezone (user): updated id=%s", id)
	return nil
}

// GetPublicByID возвращает публичные данные пользователя по UUID
func (s *Service) GetPublicByID(userID string) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	user, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Errorf("Service.GetPublicByID (user): repo error: %v", err)
		return nil, err
	}
	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *Service) GetUserByID(userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		s.logger.Errorf("Service.GetUserByID (user): repo error: %v", err)
		return nil, err
	}
	return user, nil
}

// RequestAccountDeletion отправляет запрос на подтверждение удаления в Telegram
func (s *Service) RequestAccountDeletion(userID uuid.UUID, telegramID int64) error {
	// Отправляем уведомление в Telegram с кнопками подтверждения
	err := sender.RequestAccountDeletionConfirmation(userID, telegramID)
	if err != nil {
		s.logger.Errorf("Service.RequestAccountDeletion: failed to send telegram notification: %v", err)
		return err
	}
	s.logger.Infof("Service.RequestAccountDeletion: confirmation request sent for user_id=%s", userID)
	return nil
}

// DeleteUser удаляет пользователя и все связанные данные
func (s *Service) DeleteUser(userID uuid.UUID) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		s.logger.Errorf("Service.DeleteUser: user not found %v", err)
		return nil
	}
	if err := s.repo.DeleteToken(user.TelegramID); err != nil {
		s.logger.Errorf("Service.DeleteUser: token not found: %v", err)
		return nil
	}
	if err := s.repo.DeleteUser(userID); err != nil {
		s.logger.Errorf("Service.DeleteUser: failed to delete user: %v", err)
		return err
	}

	s.logger.Infof("Service.DeleteUser: user deleted successfully user_id=%s", userID)
	return nil
}
