package role

import (
	"app/http/repository/role"
	"app/pkg/models"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	roleRepo *role.Repository
}

func NewService(roleRepo *role.Repository) *Service {
	return &Service{roleRepo: roleRepo}
}

type GetUserRolesResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []string  `json:"roles"`
}

type GetAllRolesResponse struct {
	Roles []models.UserRole `json:"roles"`
}

// CreateRole creates a new role for a user
func (s *Service) CreateRole(req models.UserRole) error {
	// Validate role name
	if req.Role == "" {
		return errors.New("role name cannot be empty")
	}

	// Validate user ID
	if req.UserID.String() == "" {
		return errors.New("invalid user ID")
	}

	// Check if role already exists
	exists, err := s.roleRepo.CheckUserRole(req.UserID, req.Role)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("user already has this role")
	}

	// Create the role
	err = s.roleRepo.CreateRole(req.UserID, req.Role)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRole removes a role from a user
func (s *Service) DeleteRole(req models.UserRole) error {
	// Validate inputs
	if req.Role == "" {
		return errors.New("role name cannot be empty")
	}

	if req.UserID.String() == "" {
		return errors.New("invalid user ID")
	}

	// Check if role exists
	exists, err := s.roleRepo.CheckUserRole(req.UserID, req.Role)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("user does not have this role")
	}

	// Delete the role
	err = s.roleRepo.DeleteRole(req.UserID, req.Role)
	if err != nil {
		return err
	}

	return nil
}

// GetUserRoles retrieves all roles for a specific user
func (s *Service) GetUserRoles(userID uuid.UUID) (*GetUserRolesResponse, error) {
	if userID.String() == "" {
		return nil, errors.New("invalid user ID")
	}

	roles, err := s.roleRepo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	return &GetUserRolesResponse{
		UserID: userID,
		Roles:  roles,
	}, nil
}

// GetAllRoles retrieves all roles in the system
func (s *Service) GetAllRoles() (*GetAllRolesResponse, error) {
	roles, err := s.roleRepo.GetAllRoles()
	if err != nil {
		return nil, err
	}

	return &GetAllRolesResponse{
		Roles: roles,
	}, nil
}

// CheckUserRole checks if a user has a specific role
func (s *Service) CheckUserRole(userID uuid.UUID, role string) (bool, error) {
	if userID.String() == "" {
		return false, errors.New("invalid user ID")
	}

	if role == "" {
		return false, errors.New("role name cannot be empty")
	}

	exists, err := s.roleRepo.CheckUserRole(userID, role)
	if err != nil {
		return false, err
	}

	return exists, nil
}
