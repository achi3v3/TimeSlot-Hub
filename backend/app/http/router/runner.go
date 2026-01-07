package router

import (
	_ "app/docs"
	metricsCtrl "app/http/controller/metrics"
	notifyCtrl "app/http/controller/notification"
	"app/http/middleware"
	metricsRepo "app/http/repository/metrics"
	notifyRepo "app/http/repository/notification"
	notifyServ "app/http/usecase/notification"
	"app/internal/database"
	"app/internal/logger"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func InternalAuthMiddleware() gin.HandlerFunc {
	allowedInternal := os.Getenv("INTERNAL_TOKEN")
	allowedFrontend := os.Getenv("FRONTEND_SECRET")
	return func(c *gin.Context) {
		// Allow claim-token via short-lived login flow cookie set after /user/login
		if strings.HasPrefix(c.Request.URL.Path, "/user/claim-token/") {
			if cookie, err := c.Cookie("login_flow"); err == nil && cookie == "1" {
				c.Next()
				return
			}
		}
		internalToken := c.GetHeader("X-Internal-Token")
		frontendToken := c.GetHeader("X-Frontend-Secret")
		// Deny only if both do not match any configured secret
		if !((allowedInternal != "" && internalToken == allowedInternal) || (allowedFrontend != "" && frontendToken == allowedFrontend)) {
			c.JSON(403, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Client) Run() error {
	// CORS middleware with secure configuration
	s.router.Use(func(c *gin.Context) {
		// Allowed domains from ALLOWED_ORIGINS environment variable (comma-separated)
		origin := c.Request.Header.Get("Origin")
		allowed := os.Getenv("ALLOWED_ORIGINS")
		if allowed == "" {
			s.logger.Warn("ALLOWED_ORIGINS not set, using defaults for development")
			allowed = "http://localhost:3000,http://localhost:5173,http://localhost:8091"
		}
		for _, o := range strings.Split(allowed, ",") {
			if strings.TrimSpace(o) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Internal-Token, X-Frontend-Secret, X-Admin-Phone")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Global security middleware
	s.router.Use(middleware.ValidateInputMiddleware())
	s.router.Use(middleware.GeneralRateLimitMiddleware())

	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "API is running")
	})
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	userHandler := s.GetUserHandler()
	userGroup := s.router.Group("/user")
	{
		// Public endpoints (no authentication required)
		userGroup.GET("/check/:telegram_id", userHandler.CheckAuth)
		userGroup.GET("/check-login/:telegram_id", userHandler.CheckLogin)
		userGroup.GET("/public/:uuid", userHandler.GetPublicUser)
		userGroup.POST("/register", userHandler.CreateUser)
		userGroup.POST("/login", userHandler.Login)

		// Internal endpoints (for Telegram bot only)
		userGroup.GET("/g3tter/:telegram_id", userHandler.GetUserByTelegramID)
		// Internal token confirmation and claim endpoints require internal token
		internalUser := userGroup.Group("")
		internalUser.Use(InternalAuthMiddleware())
		internalUser.POST("/confirm-login/:telegram_id", userHandler.ConfirmLogin)
		internalUser.POST("/claim-token/:telegram_id", userHandler.ClaimToken)
		userGroup.POST("/confirm-deletion", userHandler.ConfirmAccountDeletion)

		// Protected endpoints (require session authentication)
		userGroup.Use(middleware.SessionAuthMiddleware())
		userGroup.POST("/logout", userHandler.Logout)
		userGroup.PUT("/update", userHandler.UpdateUser)
		userGroup.PUT("/timezone", userHandler.UpdateTimezone)
		userGroup.POST("/request-deletion", userHandler.RequestAccountDeletion)
	}

	// Internal routes for Telegram bot (without user session)
	userTelegramGroup := s.router.Group("/telegram/user")
	{
		userTelegramGroup.Use(InternalAuthMiddleware())
		userTelegramGroup.PUT("/timezone", userHandler.UpdateTimezoneInternal)
	}

	slotHandler := s.GetSlotHandler()
	slotGroup := s.router.Group("/slot")
	{
		// Public endpoints (no authentication required)
		slotGroup.GET("/:uuid", slotHandler.GetSlots)
		slotGroup.GET("/one/:id", slotHandler.GetSlot)

		// Protected endpoints (require session authentication)
		slotGroup.Use(middleware.SessionAuthMiddleware())
		slotGroup.POST("/master/create", middleware.CreateRateLimitMiddleware(), slotHandler.CreateSlot)
		slotGroup.DELETE("/master/:uuid", slotHandler.DeleteSlots)
		slotGroup.DELETE("/master/one/:id", slotHandler.DeleteSlot)
	}

	recordHandler := s.GetRecordHandler()
	recordGroup := s.router.Group("/record")
	{
		// Public endpoints (no authentication required)
		recordGroup.GET("/:uuid", recordHandler.GetClientRecords)
		recordGroup.POST("/user/filter", recordHandler.GetClientRecordsFiltered)

		// Protected endpoints (require session authentication)
		recordGroup.Use(middleware.SessionAuthMiddleware())
		recordGroup.GET("/master/:slot_id", recordHandler.GetAllRecordsBySlot)
		recordGroup.POST("/master/create", recordHandler.CreateRecord)
		recordGroup.POST("/master/get", recordHandler.GetRecordsBySlot)
		recordGroup.POST("/master/status", recordHandler.UpdateRecordStatus)
		recordGroup.POST("/master/confirm/:record_id", recordHandler.ConfirmRecord)
		recordGroup.POST("/master/reject/:record_id", recordHandler.RejectRecord)
		recordGroup.DELETE("/master/:record_id", recordHandler.DeleteRecord)
	}
	recordTelegramGroup := s.router.Group("/telegram/record")
	{
		// Protected endpoints (require internal authentication)
		recordTelegramGroup.Use(InternalAuthMiddleware())
		recordTelegramGroup.POST("/master/create", recordHandler.CreateRecord)
		recordTelegramGroup.POST("/master/status", recordHandler.UpdateRecordStatus)
		recordTelegramGroup.POST("/master/confirm/:record_id", recordHandler.ConfirmRecord)
		recordTelegramGroup.GET("/master/upcoming/:telegram_id", recordHandler.GetUpcomingRecordsByMasterTelegramID)
	}
	tokenMap := &sync.Map{}
	serviceHandler := s.GetServiceHandler(tokenMap)
	serviceGroup := s.router.Group("/service")
	{
		// Public endpoints (no authentication required)
		serviceGroup.GET("/master/:uuid", serviceHandler.GetServices)
		serviceGroup.GET("/:id", serviceHandler.GetService)

		// Protected endpoints (require session authentication)
		serviceGroup.Use(middleware.SessionAuthMiddleware())
		serviceGroup.POST("/create", serviceHandler.CreateService)
		serviceGroup.PUT("/update", serviceHandler.UpdateService)
		serviceGroup.DELETE("/:id", serviceHandler.DeleteService)
	}

	// Notification routes
	notifyRepo := notifyRepo.NewRepository(s.gormDB, s.logger)
	notifyServ := notifyServ.NewService(notifyRepo, s.logger)
	notifyHandler := notifyCtrl.NewHandler(notifyServ, s.logger)
	notifyGroup := s.router.Group("/notification")
	{
		// Notifications require authentication
		notifyGroup.Use(middleware.SessionAuthMiddleware())
		notifyGroup.GET("/", notifyHandler.GetClientNotifications)
		notifyGroup.GET("/unread-count", notifyHandler.CountUnreadUserNotifications)
		notifyGroup.POST("/:id/mark-read", notifyHandler.MarkIsReadUserNotification)
		notifyGroup.POST("/mark-all-read", notifyHandler.MarkReadAllUserNotifications)
	}

	// Admin routes
	adminLogger := &logger.Logger{Logger: s.logger}
	dbStruct := &database.Database{DB: s.gormDB}
	SetupAdminRoutes(s.router, dbStruct, adminLogger, s.GetRoleHandler(), notifyServ)

	// Metrics routes (public, rate-limited globally)
	mrepo := metricsRepo.NewRepository(s.gormDB, s.logger)
	mhandler := metricsCtrl.NewHandler(mrepo)
	s.router.POST("/metrics/ad-click", mhandler.TrackAdClick)

	// Swagger documentation
	s.logger.Infof("API documentation available at: http://localhost:8090/swagger/index.html#/")
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	s.logger.Infof("Starting server on :8090")
	if err := s.httpServer.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("Server failse: %v", err)
	}

	s.logger.Info("Server stopped")
	return nil
}
