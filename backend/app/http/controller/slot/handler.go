package slot

import (
	"app/http/utils"
	"app/pkg/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSlot creates a new time slot
// @Summary Create slot
// @Description Create a new time slot for master
// @Tags slot
// @Accept json
// @Produce json
// @Param slot body models.Slot true "Slot data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /slot/master/create [post]
func (h *Handler) CreateSlot(ctx *gin.Context) {
	var slot models.Slot
	if err := ctx.ShouldBindJSON(&slot); err != nil {
		h.logger.WithError(err).Error("CreateSlot: invalid request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data. Please check date and time."})
		return
	}
	if slot.MasterID == (uuid.UUID{}) {
		h.logger.Error("CreateSlot: master_id is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "master_id is required"})
		return
	}

	// Validate slot data
	if slot.StartTime.IsZero() || slot.EndTime.IsZero() {
		h.logger.Errorf("CreateSlot: start time or end time is zero, master_id: %s", slot.MasterID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Start time and end time are required"})
		return
	}

	if slot.EndTime.Before(slot.StartTime) {
		h.logger.Errorf("CreateSlot: end time is before start time, master_id: %s, start_time: %v, end_time: %v", slot.MasterID.String(), slot.StartTime, slot.EndTime)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	// Service is required: prevent DB errors and schema details leakage
	if slot.ServiceID == 0 {
		h.logger.Errorf("CreateSlot: service_id is missing, master_id: %s", slot.MasterID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Service must be selected before creating slot"})
		return
	}

	if err := h.service.CreateSlot(&slot); err != nil {
		h.logger.Errorf("CreateSlot: failed to create slot in service layer: %v, master_id: %s, service_id: %d, start_time: %v, end_time: %v", err, slot.MasterID.String(), slot.ServiceID, slot.StartTime, slot.EndTime)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create slot. Please check selected service and time."})
		return
	}

	h.logger.Infof("CreateSlot: slot created successfully, slot_id: %d, master_id: %s, service_id: %d", slot.ID, slot.MasterID.String(), slot.ServiceID)

	ctx.JSON(http.StatusOK, gin.H{"message": "Slot created"})
}

// GetSlots returns slots by master uuid
// @Summary Get slots
// @Description Get slots by master uuid
// @Tags slot
// @Produce json
// @Param uuid path string true "Master UUID"
// @Success 200 {array} models.Slot
// @Failure 400 {object} map[string]string
// @Router /slot/{uuid} [get]
func (h *Handler) GetSlots(ctx *gin.Context) {
	param := ctx.Param("uuid")
	userID, err := uuid.Parse(param)
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: invalid master_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	slotsWithMaster, err := h.service.GetSlots(userID)
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slots for master_id=%v", userID)
	ctx.JSON(http.StatusOK, slotsWithMaster)
}

// GetSlot returns slot by id
// @Summary Get slot
// @Description Get slot by ID
// @Tags slot
// @Produce json
// @Param id path string true "Slot ID"
// @Success 200 {object} models.Slot
// @Failure 400 {object} map[string]string
// @Router /slot/one/{id} [get]
func (h *Handler) GetSlot(ctx *gin.Context) {
	param := ctx.Param("id")
	slotID, err := strconv.ParseUint(param, 10, 0)
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: invalid master_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	slotsResponce, err := h.service.GetSlot(uint(slotID))
	fmt.Println(slotsResponce)
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slot by id=%v", slotID)
	ctx.JSON(http.StatusOK, slotsResponce)
}

// DeleteSlots deletes all slots for master
// @Summary Delete slots by master
// @Description Delete all slots for master uuid
// @Tags slot
// @Produce json
// @Param uuid path string true "Master UUID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /slot/master/{uuid} [delete]
func (h *Handler) DeleteSlots(ctx *gin.Context) {
	param := ctx.Param("uuid")
	requestedUserID, err := uuid.Parse(param)
	if err != nil {
		h.logger.Errorf("Handler.DeleteSlots: invalid master_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Owner check is disabled in session-based scheme; valid UUID is sufficient

	err = h.service.DeleteSlots(requestedUserID)
	if err != nil {
		h.logger.Errorf("Handler.DeleteSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Slots deleted"})
}

// DeleteSlot deletes slot with owner verification
// @Summary Delete slot (owner)
// @Description Delete single slot with owner verification
// @Tags slot
// @Produce json
// @Param id path string true "Slot ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /slot/master/one/{id} [delete]
func (h *Handler) DeleteSlot(ctx *gin.Context) {
	// Extract user_id from token
	userID, err := utils.ExtractUserIDFromToken(ctx)
	if err != nil {
		h.logger.Errorf("DeleteSlot: auth error: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.logger.Errorf("DeleteSlot: invalid slot id: %v, param: %s", err, ctx.Param("id"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	// Delete slot with owner verification
	err = h.service.DeleteSlotByOwner(uint(id), userID)
	if err != nil {
		h.logger.Errorf("DeleteSlot: service error: %v, slot_id: %d, user_id: %s", err, id, userID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("DeleteSlot: slot deleted successfully, slot_id: %d, user_id: %s", id, userID.String())
	ctx.JSON(http.StatusOK, gin.H{"message": "Slot deleted"})
}
