package admin

import (
	"app/pkg/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserDetail returns detailed user information
// @Summary Get user details
// @Description Get detailed information about user including slots, services, and records
// @Tags admin
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/user/{id} [get]
func (h *Handler) GetUserDetail(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.logger.Errorf("GetUserDetail: invalid user ID format: %v, param: %s", err, ctx.Param("id"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		h.logger.Errorf("GetUserDetail: failed to get user from database: %v, user_id: %s", err, userID.String())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		h.logger.Errorf("GetUserDetail: user not found, user_id: %s", userID.String())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	roles, err := h.userRepo.GetUserRoles(userID)
	if err != nil {
		h.logger.Errorf("GetUserDetail: failed to get user roles, using empty array: %v, user_id: %s", err, userID.String())
		roles = []string{}
	}

	slots, err := h.slotRepo.GetSlotsByMasterID(userID)
	if err != nil {
		h.logger.Errorf("GetUserDetail: failed to get user slots, using empty array: %v, user_id: %s", err, userID.String())
		slots = []models.Slot{}
	} else {
		h.logger.Infof("GetUserDetail: retrieved user slots, user_id: %s, slots_count: %d", userID.String(), len(slots))
	}

	services, err := h.serviceRepo.GetServicesByMasterID(userID)
	if err != nil {
		h.logger.Errorf("GetUserDetail: failed to get user services, using empty array: %v, user_id: %s", err, userID.String())
		services = []models.Service{}
	} else {
		h.logger.Infof("GetUserDetail: retrieved user services, user_id: %s, services_count: %d", userID.String(), len(services))
	}

	records, err := h.recordRepo.GetRecordsByMasterID(userID)
	if err != nil {
		h.logger.Errorf("GetUserDetail: failed to get user records, using empty array: %v, user_id: %s", err, userID.String())
		records = []models.Record{}
	} else {
		h.logger.Infof("GetUserDetail: retrieved user records, user_id: %s, records_count: %d", userID.String(), len(records))
	}
	userInfo := UserInfo{
		ID:         user.ID.String(),
		FirstName:  user.FirstName,
		Surname:    user.Surname,
		Phone:      user.Phone,
		TelegramID: user.TelegramID,
		IsActive:   user.Active,
		Roles:      roles,
		CreatedAt:  user.ConsentGivenAt,
	}

	response := UserDetailResponse{
		User:     userInfo,
		Slots:    slots,
		Services: services,
		Records:  records,
	}

	h.logger.Infof("GetUserDetail: sending response, user_id: %s, slots_count: %d, services_count: %d, records_count: %d",
		userID.String(), len(slots), len(services), len(records))

	ctx.JSON(http.StatusOK, response)
}

// GetDetailSlot returns detailed slot information
// @Summary Get slot details
// @Description Get detailed information about a slot
// @Tags admin
// @Produce json
// @Param id path string true "Slot ID"
// @Success 200 {object} models.Slot
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/slot/{id} [get]
func (h *Handler) GetDetailSlot(ctx *gin.Context) {
	slotIDStr := ctx.Param("id")
	slotID, err := strconv.ParseUint(slotIDStr, 10, 32)
	if err != nil {
		h.logger.Errorf("GetDetailSlot: invalid slot ID format: %v, param: %s", err, slotIDStr)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	detailSlot, err := h.slotServ.GetSlot(uint(slotID))
	if err != nil {
		h.logger.Errorf("GetDetailSlot: failed to get slot from service: %v, slot_id: %d", err, slotID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slot"})
		return
	}

	h.logger.Infof("GetDetailSlot: slot retrieved successfully, slot_id: %d", slotID)
	ctx.JSON(http.StatusOK, detailSlot)
}

// GetDetailService returns detailed service information
// @Summary Get service details
// @Description Get detailed information about a service
// @Tags admin
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} models.Service
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/service/{id} [get]
func (h *Handler) GetDetailService(ctx *gin.Context) {
	serviceIDStr := ctx.Param("id")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		h.logger.Errorf("GetDetailService: invalid service ID format: %v, param: %s", err, serviceIDStr)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	detailService, err := h.serviceServ.GetDetailService(uint(serviceID))
	if err != nil {
		h.logger.Errorf("GetDetailService: failed to get service from service layer: %v, service_id: %d", err, serviceID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service"})
		return
	}

	h.logger.Infof("GetDetailService: service retrieved successfully, service_id: %d", serviceID)
	ctx.JSON(http.StatusOK, detailService)
}

// GetDetailRecord returns detailed record information
// @Summary Get record details
// @Description Get detailed information about a record
// @Tags admin
// @Produce json
// @Param id path string true "Record ID"
// @Success 200 {object} models.Record
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/record/{id} [get]
func (h *Handler) GetDetailRecord(ctx *gin.Context) {
	recordIDStr := ctx.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		h.logger.Errorf("GetDetailRecord: invalid record ID format: %v, param: %s", err, recordIDStr)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	detailRecord, err := h.recordServ.GetDetailRecord(uint(recordID))
	if err != nil {
		h.logger.Errorf("GetDetailRecord: failed to get record from service layer: %v, record_id: %d", err, recordID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get record"})
		return
	}

	h.logger.Infof("GetDetailRecord: record retrieved successfully, record_id: %d", recordID)
	ctx.JSON(http.StatusOK, detailRecord)
}
