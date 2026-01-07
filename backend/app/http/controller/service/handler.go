package service

import (
	"app/http/utils"
	"app/pkg/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateService creates a new service
// @Summary Create service
// @Description Create a new service for master
// @Tags service
// @Accept json
// @Produce json
// @Param service body models.Service true "Service data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /service/create [post]
func (h *Handler) CreateService(ctx *gin.Context) {
	var service models.Service
	if err := ctx.ShouldBindJSON(&service); err != nil {
		h.logger.Errorf("Handler.CreateService: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Если master_id не передан во входных данных — возвращаем ошибку
	if service.MasterID == (uuid.UUID{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "master_id is required"})
		return
	}

	// Валидация данных услуги
	if len(service.Name) < 1 || len(service.Name) > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Service name must be between 1 and 100 characters"})
		return
	}

	if service.Price < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be negative"})
		return
	}

	if service.Duration <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Duration must be positive"})
		return
	}

	if err := h.service.CreateService(&service); err != nil {
		h.logger.Errorf("Handler.CreateService: create service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Service created successfully",
		"result":  true,
	})
}

// GetServices returns services for master (by uuid or telegram_id)
// @Summary Get services
// @Description Get services for master by uuid or telegram_id
// @Tags service
// @Produce json
// @Param uuid path string true "Master UUID or telegram_id"
// @Success 200 {array} models.Service
// @Failure 400 {object} map[string]string
// @Router /service/master/{uuid} [get]
func (h *Handler) GetServices(ctx *gin.Context) {
	param := ctx.Param("uuid")
	userID, err := uuid.Parse(param)
	if err != nil {
		// Если не UUID, то это может быть telegram_id - нужно найти master_id
		h.logger.Infof("Handler.GetServices: treating as telegram_id: %s", param)
		// Пытаемся найти пользователя по telegram_id
		telegramID, parseErr := strconv.ParseInt(param, 10, 64)
		if parseErr != nil {
			h.logger.Errorf("Handler.GetServices: invalid parameter format: %s", param)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter format"})
			return
		}

		// Ищем пользователя по telegram_id и получаем его услуги
		services, err := h.service.GetServicesByTelegramID(telegramID)
		if err != nil {
			h.logger.Errorf("Handler.GetServices: telegram_id lookup failed: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			return
		}
		h.logger.Infof("Handler.GetServices: sending services for telegram_id=%d", telegramID)
		ctx.JSON(http.StatusOK, services)
		return
	}

	services, err := h.service.GetServices(userID)
	if err != nil {
		h.logger.Errorf("Handler.GetServices: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetServices: sending services for master_id=%v", userID)
	ctx.JSON(http.StatusOK, services)
}

// GetService returns service by id
// @Summary Get service
// @Description Get service by ID
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} models.Service
// @Failure 400 {object} map[string]string
// @Router /service/{id} [get]
func (h *Handler) GetService(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.GetService: invalid master_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	service, err := h.service.GetService(uint(id))
	if err != nil {
		h.logger.Errorf("Handler.GetService: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetService: sending service with id=%d", id)
	ctx.JSON(http.StatusOK, service)
}

// UpdateService updates service
// @Summary Update service
// @Description Update service fields
// @Tags service
// @Accept json
// @Produce json
// @Param service body models.Service true "Service update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /service/update [put]
func (h *Handler) UpdateService(ctx *gin.Context) {
	var service models.Service
	if err := ctx.ShouldBindJSON(&service); err != nil {
		h.logger.Errorf("Handler.UpdateService: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	if service.MasterID == (uuid.UUID{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "master_id is required"})
		return
	}

	// Валидация данных услуги
	if len(service.Name) < 1 || len(service.Name) > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Service name must be between 1 and 100 characters"})
		return
	}

	if service.Price < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be negative"})
		return
	}

	if service.Duration <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Duration must be positive"})
		return
	}

	if err := h.service.UpdateService(&service); err != nil {
		h.logger.Errorf("Handler.UpdateService: update error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Service updated successfully",
		"result":  true,
	})
}

// DeleteService deletes service with owner check
// @Summary Delete service
// @Description Delete service by ID (owner check)
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /service/{id} [delete]
func (h *Handler) DeleteService(ctx *gin.Context) {
	// Извлекаем user_id из токена
	userID, err := utils.ExtractUserIDFromToken(ctx)
	if err != nil {
		h.logger.Errorf("Handler.DeleteService: auth error: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.DeleteService: invalid service id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Удаляем услугу с проверкой владельца
	err = h.service.DeleteServiceByOwner(uint(id), userID)
	if err != nil {
		h.logger.Errorf("Handler.DeleteService: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Service deleted"})
}
