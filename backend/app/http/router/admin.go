package router

import (
	"app/http/controller/admin"
	"app/http/controller/role"
	"app/http/middleware"
	mrepo "app/http/repository/metrics"
	"app/http/repository/record"
	"app/http/repository/service"
	"app/http/repository/slot"
	"app/http/repository/user"
	"app/http/usecase/notification"
	recordServ "app/http/usecase/record"
	serviceServ "app/http/usecase/service"
	slotServ "app/http/usecase/slot"
	userServ "app/http/usecase/user"
	"app/internal/database"
	"app/internal/logger"
	"sync"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes настраивает маршруты для админки
func SetupAdminRoutes(r *gin.Engine, db *database.Database, logger *logger.Logger, roleHandler *role.Handler, notifyServ *notification.Service) {
	// Создаем репозитории с logrus.Logger
	logrusLogger := logger.Logger
	tokenMap := &sync.Map{}
	userRepo := user.NewRepository(db.DB, logrusLogger, tokenMap)
	slotRepo := slot.NewRepository(db.DB, logrusLogger)
	serviceRepo := service.NewRepository(db.DB, logrusLogger)
	recordRepo := record.NewRepository(db.DB, logrusLogger)
	metricsRepo := mrepo.NewRepository(db.DB, logrusLogger)

	recordService := recordServ.NewService(recordRepo, notifyServ, logrusLogger)
	serviceService := serviceServ.NewService(serviceRepo, userRepo, logrusLogger)
	slotService := slotServ.NewService(slotRepo, logrusLogger)
	userService := userServ.NewService(userRepo, logrusLogger)
	// Создаем админский хендлер
	adminHandler := admin.NewHandler(userRepo, slotRepo, serviceRepo, recordRepo, metricsRepo, userService, slotService, serviceService, recordService, logger)

	// Группа маршрутов для админки
	adminGroup := r.Group("/admin")
	{
		// Публичные маршруты (требуют только номер телефона)
		adminGroup.POST("/login", adminHandler.AdminLogin)

		// Защищенные маршруты (требуют админскую авторизацию)
		protected := adminGroup.Group("")
		protected.Use(middleware.AdminOrPhoneMiddleware())
		{
			// Статистика
			protected.GET("/stats", adminHandler.GetStats)

			// Управление пользователями
			protected.GET("/users", adminHandler.GetUsers)
			protected.GET("/users/:id", adminHandler.GetUserDetail)
			protected.DELETE("/users/:id", adminHandler.DeleteUser)
			protected.POST("/users/:id/toggle-active", adminHandler.ToggleUserActive)

			// Роли пользователей - админские маршруты
			protected.POST("/roles", roleHandler.CreateRole)
			protected.DELETE("/roles", roleHandler.DeleteRole)
			protected.GET("/roles", roleHandler.GetAllRoles)

			protected.GET("/slots", adminHandler.GetAllSlots)
			protected.DELETE("/slots/:id", adminHandler.DeleteSlot)
			protected.GET("/slots/:id", adminHandler.GetDetailSlot)
			protected.GET("/services/:id", adminHandler.GetDetailService)
			protected.GET("/records/:id", adminHandler.GetDetailRecord)
			protected.GET("/services", adminHandler.GetAllServices)
			protected.GET("/records", adminHandler.GetAllRecords)

			// Роли конкретного пользователя
			adminUserRoleGroup := protected.Group("/users/:id/roles")
			{
				adminUserRoleGroup.GET("/", roleHandler.GetUserRoles)
				adminUserRoleGroup.GET("/:role", roleHandler.CheckUserRole)
			}
		}
	}
}
