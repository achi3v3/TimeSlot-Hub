package role

import (
	roleRepo "app/http/repository/role"
	roleServ "app/http/usecase/role"

	"gorm.io/gorm"
)

type Handler struct {
	roleService *roleServ.Service
}

func NewHandler(db *gorm.DB) *Handler {
	roleRepo := roleRepo.NewRepository(db)
	roleService := roleServ.NewService(roleRepo)

	return &Handler{
		roleService: roleService,
	}
}
