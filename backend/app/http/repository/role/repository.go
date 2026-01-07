package role

import (
	"app/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateRole creates a new role for a user
func (r *Repository) CreateRole(userID uuid.UUID, roleName string) error {
	role := models.UserRole{
		UserID: userID,
		Role:   roleName,
	}

	// Use GORM's FirstOrCreate to avoid duplicates
	result := r.db.Where("user_id = ? AND role = ?", userID, roleName).FirstOrCreate(&role)
	return result.Error
}

// DeleteRole removes a role from a user
func (r *Repository) DeleteRole(userID uuid.UUID, roleName string) error {
	result := r.db.Where("user_id = ? AND role = ?", userID, roleName).Delete(&models.UserRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetUserRoles retrieves all roles for a specific user
func (r *Repository) GetUserRoles(userID uuid.UUID) ([]string, error) {
	var roles []models.UserRole
	err := r.db.Where("user_id = ?", userID).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Role)
	}

	return roleNames, nil
}

// GetAllRoles retrieves all roles in the system
func (r *Repository) GetAllRoles() ([]models.UserRole, error) {
	var roles []models.UserRole
	err := r.db.Find(&roles).Error
	return roles, err
}

// CheckUserRole checks if a user has a specific role
func (r *Repository) CheckUserRole(userID uuid.UUID, roleName string) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserRole{}).Where("user_id = ? AND role = ?", userID, roleName).Count(&count).Error
	return count > 0, err
}
