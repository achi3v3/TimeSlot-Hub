package user

import (
	"app/pkg/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create user
func (r *Repository) Create(user *models.User) error {
	user.ConsentGivenAt = time.Now()
	user.PrivacyPolicyAcceptedAt = time.Now()
	user.TermsAcceptedAt = time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Roles").Create(user).Error; err != nil {
			r.logger.Errorf("Repository.Create (user): create failed: %v", err)
			return err
		}
		if err := tx.Create(&models.UserRole{
			UserID: user.ID,
			Role:   "client",
		}).Error; err != nil {
			r.logger.Errorf("Repository.Create (user): role create failed: %v", err)
			return err
		}
		r.logger.Infof("Repository.Create (user): created id=%s", user.ID)
		return nil
	})
}

// Create role at user
func (r *Repository) SetRole(userID uuid.UUID, role string) error {

	var count int64
	var userRole models.UserRole
	r.db.Model(&userRole).
		Where("user_id = ? AND role = ?", userID, role).
		Count(&count)

	if count > 0 {
		r.logger.Infof("Repository.SetRole (user): role already exists user_id=%s role=%s", userID, role)
		return nil
	}

	newRole := models.UserRole{
		UserID: userID,
		Role:   role,
	}
	if err := r.db.Create(&newRole).Error; err != nil {
		r.logger.Errorf("Repository.SetRole (user): create failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.SetRole (user): role created user_id=%s role=%s", userID, role)
	return nil

}

// Find user by telegram ID
func (r *Repository) FindByTelegramID(telegram_id int64) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Roles").
		Preload("Services").
		Where("telegram_id = ?", telegram_id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Infof("Repository.FindByTelegramID (user): not found telegram_id=%d", telegram_id)
		return nil, nil
	}
	if err != nil {
		r.logger.Errorf("Repository.FindByTelegramID (user): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindByTelegramID (user): found id=%s", user.ID)
	return &user, err
}

type tempToken struct {
	Value     string
	CreatedAt time.Time
}

const tokenTTL = time.Hour

// Store user token by telegram ID at KV-repository
func (r *Repository) StorageToken(telegram_id int64, token string) error {
	r.mapToken.Store(telegram_id, tempToken{Value: token, CreatedAt: time.Now()})
	return nil
}
func (r *Repository) DeleteToken(telegram_id int64) error {
	r.mapToken.Delete(telegram_id)
	return nil
}

// Claim user token from KV-repository by key telegram ID
func (r *Repository) ClaimUserToken(telegram_id int64) (string, error) {
	v, ok := r.mapToken.Load(telegram_id)
	if !ok {
		r.logger.Infof("Repository.LoadUserToken (user): not found telegram_id=%d", telegram_id)
		return "", nil
	}
	// Ensure not expired
	tt, isStruct := v.(tempToken)
	if isStruct {
		if time.Since(tt.CreatedAt) > tokenTTL {
			r.mapToken.Delete(telegram_id)
			r.logger.Infof("Repository.LoadUserToken (user): expired telegram_id=%d", telegram_id)
			return "", nil
		}
		r.mapToken.Delete(telegram_id)
		r.logger.Infof("Repository.LoadUserToken (user): found telegram_id=%d", telegram_id)
		return tt.Value, nil
	}
	// Backward compatibility if old string was stored
	r.mapToken.Delete(telegram_id)
	r.logger.Infof("Repository.LoadUserToken (user): found (legacy) telegram_id=%d", telegram_id)
	return v.(string), nil
}

// Check exists user token at KV-repository by telegram ID
func (r *Repository) CheckUserToken(telegram_id int64) bool {
	v, ok := r.mapToken.Load(telegram_id)
	if !ok {
		return false
	}
	if tt, isStruct := v.(tempToken); isStruct {
		if time.Since(tt.CreatedAt) > tokenTTL {
			r.mapToken.Delete(telegram_id)
			return false
		}
		return true
	}
	// Legacy value without ttl; consider valid for compatibility
	return true
}

// Find user by phone
func (r *Repository) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Roles").
		Where("phone = ?", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Infof("Repository.FindByPhone (user): not found phone=%s", phone)
		return nil, nil
	}
	if err != nil {
		r.logger.Errorf("Repository.FindByPhone (user): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindByPhone (user): found id=%s", user.ID)
	return &user, nil
}

// Find user by uuid
func (r *Repository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.
		Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Infof("Repository.FindByID (user): not found id=%s", id)
		return nil, nil
	}
	if err != nil {
		r.logger.Errorf("Repository.FindByID (user): query failed: %v", err)
		return nil, err
	}
	r.logger.Infof("Repository.FindByID (user): found id=%s", user.ID)
	return &user, nil
}

// Update fields(firstname and surname) by user uuid
func (r *Repository) UpdateNames(userID uuid.UUID, firstName string, surname string) error {
	// Обновляем только допустимые поля
	if err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"first_name": firstName,
			"surname":    surname,
		}).Error; err != nil {
		r.logger.Errorf("Repository.UpdateNames (user): update failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.UpdateNames (user): updated id=%s", userID)
	return nil
}

// Update user field timezone
func (r *Repository) UpdateTimezone(userID uuid.UUID, timezone string) error {
	if timezone == "" {
		return nil
	}
	if err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("timezone", timezone).Error; err != nil {
		r.logger.Errorf("Repository.UpdateTimezone (user): update failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.UpdateTimezone (user): updated id=%s", userID)
	return nil
}

// Delete user by uuid
func (r *Repository) DeleteUser(userID uuid.UUID) error {
	err := r.db.Where("id = ?", userID).Delete(&models.User{}).Error
	if err != nil {
		r.logger.Errorf("Repository.DeleteUser (user): delete failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.DeleteUser (user): deleted id=%s", userID)
	return nil
}

// Count all users
func (r *Repository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountUsers: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// Count all active users
func (r *Repository) CountActiveUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("active = ?", true).Count(&count).Error
	if err != nil {
		r.logger.Errorf("Repository.CountActiveUsers: count failed: %v", err)
		return 0, err
	}
	return count, nil
}

// Return users with pagination by page and limit
func (r *Repository) GetUsersWithPagination(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	err := r.db.Model(&models.User{}).Count(&total).Error
	if err != nil {
		r.logger.Errorf("Repository.GetUsersWithPagination: count failed: %v", err)
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.
		Preload("Roles").
		Order("consent_given_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error
	if err != nil {
		r.logger.Errorf("Repository.GetUsersWithPagination: query failed: %v", err)
		return nil, 0, err
	}

	return users, total, nil
}

// Get roles by user throug uuid
func (r *Repository) GetUserRoles(userID uuid.UUID) ([]string, error) {
	var roles []models.UserRole
	err := r.db.Where("user_id = ?", userID).Find(&roles).Error
	if err != nil {
		r.logger.Errorf("Repository.GetUserRoles: query failed: %v", err)
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Role
	}

	return roleNames, nil
}

// Return ALL user information by uuid
func (r *Repository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Roles").
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		r.logger.Errorf("Repository.GetUserByID: query failed: %v", err)
		return nil, err
	}
	return &user, nil
}

// Set written status of field active by user uuid
func (r *Repository) UpdateUserActive(userID uuid.UUID, active bool) error {
	err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("active", active).Error
	if err != nil {
		r.logger.Errorf("Repository.UpdateUserActive: update failed: %v", err)
		return err
	}
	r.logger.Infof("Repository.UpdateUserActive: updated id=%s, active=%t", userID, active)
	return nil
}
