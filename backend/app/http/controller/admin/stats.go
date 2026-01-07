package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStats returns overall system statistics
// @Summary Get system statistics
// @Description Get overall statistics including users, slots, records, and services
// @Tags admin
// @Produce json
// @Success 200 {object} AdminStatsResponse
// @Failure 500 {object} map[string]string
// @Router /admin/stats [get]
func (h *Handler) GetStats(ctx *gin.Context) {
	totalUsers, err := h.userRepo.CountUsers()
	if err != nil {
		h.logger.Errorf("GetStats: failed to count users: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user statistics"})
		return
	}

	activeUsers, err := h.userRepo.CountActiveUsers()
	if err != nil {
		h.logger.Errorf("GetStats: failed to count active users: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active user statistics"})
		return
	}

	totalSlots, err := h.slotRepo.CountSlots()
	if err != nil {
		h.logger.Errorf("GetStats: failed to count slots: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slot statistics"})
		return
	}

	bookedSlots, err := h.slotRepo.CountBookedSlots()
	if err != nil {
		h.logger.Errorf("GetStats: failed to count booked slots: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get booked slot statistics"})
		return
	}

	totalRecords, err := h.recordRepo.CountRecords()
	if err != nil {
		h.logger.Errorf("GetStats: failed to count records: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get record statistics"})
		return
	}

	pendingRecords, err := h.recordRepo.CountRecordsByStatus("pending")
	if err != nil {
		h.logger.Errorf("Handler.GetStats: failed to count pending records: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending record statistics"})
		return
	}

	confirmedRecords, err := h.recordRepo.CountRecordsByStatus("confirm")
	if err != nil {
		h.logger.Errorf("Handler.GetStats: failed to count confirmed records: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get confirmed record statistics"})
		return
	}

	rejectedRecords, err := h.recordRepo.CountRecordsByStatus("reject")
	if err != nil {
		h.logger.Errorf("Handler.GetStats: failed to count rejected records: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rejected record statistics"})
		return
	}

	totalServices, err := h.serviceRepo.CountServices()
	if err != nil {
		h.logger.Errorf("Handler.GetStats: failed to count services: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service statistics"})
		return
	}

	// Ad click stats (best-effort; ignore errors)
	var ad1, ad2 int64
	if h.metricsRepo != nil {
		if c1, c2, err := h.metricsRepo.GetTotals(); err == nil {
			ad1, ad2 = c1, c2
		}
	}

	stats := AdminStatsResponse{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		TotalSlots:       totalSlots,
		BookedSlots:      bookedSlots,
		TotalRecords:     totalRecords,
		PendingRecords:   pendingRecords,
		ConfirmedRecords: confirmedRecords,
		RejectedRecords:  rejectedRecords,
		TotalServices:    totalServices,
		AdClicks1:        ad1,
		AdClicks2:        ad2,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now().Unix(),
	})
}

// GetUsers returns list of all users with pagination
// @Summary Get users list
// @Description Get paginated list of all users
// @Tags admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} UserListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users [get]
func (h *Handler) GetUsers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.userRepo.GetUsersWithPagination(page, limit)
	if err != nil {
		h.logger.Errorf("Handler.GetUsers: failed to get users: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	userInfos := make([]UserInfo, len(users))
	for i, user := range users {
		roles := make([]string, len(user.Roles))
		for j, role := range user.Roles {
			roles[j] = role.Role
		}

		userInfos[i] = UserInfo{
			ID:         user.ID.String(),
			FirstName:  user.FirstName,
			Surname:    user.Surname,
			Phone:      user.Phone,
			TelegramID: user.TelegramID,
			IsActive:   user.Active,
			Roles:      roles,
			CreatedAt:  user.ConsentGivenAt,
		}
	}

	response := UserListResponse{
		Users: userInfos,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetAllSlots returns all slots (admin)
// @Summary Get all slots
// @Description Get all slots with masters (admin)
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/slots [get]
func (h *Handler) GetAllSlots(ctx *gin.Context) {
	slotsWithMaster, err := h.slotRepo.FindAllSlots()
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slots")
	ctx.JSON(http.StatusOK, slotsWithMaster)
}

// GetAllServices returns all services (admin)
// @Summary Get all services
// @Description Get all services (admin)
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/services [get]
func (h *Handler) GetAllServices(ctx *gin.Context) {
	services, err := h.serviceRepo.GetAllServices()
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slots")
	ctx.JSON(http.StatusOK, services)
}

// GetAllRecords returns all records (admin)
// @Summary Get all records
// @Description Get all records (admin)
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /admin/records [get]
func (h *Handler) GetAllRecords(ctx *gin.Context) {
	records, err := h.recordRepo.GetAllRecords()
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slots")
	ctx.JSON(http.StatusOK, records)
}
