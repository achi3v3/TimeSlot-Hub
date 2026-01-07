package notification

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetClientNotifications returns notifications for authenticated user
// @Summary Get notifications
// @Description Get notifications for authenticated user
// @Tags notification
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /notification/ [get]
func (h *Handler) GetClientNotifications(ctx *gin.Context) {
	userUUIDInterface, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("Handler.GetClientNotifications: user_id not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userUUIDInterface.(uuid.UUID)
	if !ok {
		h.logger.Errorf("Handler.GetClientNotifications: invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notifications, err := h.service.GetUserNotifications(userUUID)
	if err != nil {
		h.logger.Errorf("Handler.GetClientNotifications: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered book", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    notifications,
	})
}

// CountUnreadUserNotifications returns unread notifications count
// @Summary Count unread notifications
// @Description Count unread notifications for current user
// @Tags notification
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /notification/unread-count [get]
func (h *Handler) CountUnreadUserNotifications(ctx *gin.Context) {
	userUUIDInterface, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("Handler.CountUnreadUserNotifications: user_id not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userUUIDInterface.(uuid.UUID)
	if !ok {
		h.logger.Errorf("Handler.CountUnreadUserNotifications: invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	countNotifications, err := h.service.CountUserNotifications(userUUID)
	if err != nil {
		h.logger.Errorf("Handler.CountUnreadUserNotifications: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered book", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"count":   countNotifications,
	})
}

// MarkIsReadUserNotification marks single notification as read
// @Summary Mark notification read
// @Description Mark specific notification as read
// @Tags notification
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notification/{id}/mark-read [post]
func (h *Handler) MarkIsReadUserNotification(ctx *gin.Context) {
	userUUIDInterface, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("Handler.MarkIsReadUserNotification: user_id not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userUUIDInterface.(uuid.UUID)
	if !ok {
		h.logger.Errorf("Handler.MarkIsReadUserNotification: invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 0)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.service.MarkIsReadNotification(uint(id), userUUID, true)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Notification not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// MarkReadAllUserNotifications marks all notifications as read
// @Summary Mark all notifications read
// @Description Mark all notifications as read for current user
// @Tags notification
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notification/mark-all-read [post]
func (h *Handler) MarkReadAllUserNotifications(ctx *gin.Context) {
	userUUIDInterface, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("Handler.MarkReadAllUserNotifications: user_id not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userUUIDInterface.(uuid.UUID)
	if !ok {
		h.logger.Errorf("Handler.MarkReadAllUserNotifications: invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.service.MarkAllReadNotifications(userUUID)
	if err != nil {
		h.logger.Errorf("Handler.MarkReadAllUserNotifications: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered book", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}
