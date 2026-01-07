package record

import (
	"app/pkg/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRecord creates a new booking record
// @Summary Create record
// @Description Create a new booking record for a slot
// @Tags record
// @Accept json
// @Produce json
// @Param record body models.Record true "Record data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /record/master/create [post]
func (h *Handler) CreateRecord(ctx *gin.Context) {
	var book models.Record
	if err := ctx.ShouldBindJSON(&book); err != nil {
		h.logger.Errorf("Handler.CreateRecord: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate record data
	if book.SlotID == 0 {
		h.logger.Errorf("CreateRecord: slot ID is missing, client_id: %s", book.ClientID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Slot ID is required"})
		return
	}

	if book.ClientID == (uuid.UUID{}) {
		h.logger.Errorf("CreateRecord: client ID is missing or invalid UUID, slot_id: %d", book.SlotID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	if err := h.service.Create(&book); err != nil {
		h.logger.Errorf("CreateRecord: failed to create record in service layer: %v, slot_id: %d, client_id: %s", err, book.SlotID, book.ClientID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	h.logger.Infof("CreateRecord: record created successfully, record_id: %d, slot_id: %d, client_id: %s, status: %s", book.ID, book.SlotID, book.ClientID.String(), book.Status)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Record successfully created",
		"data":    book,
	})
}

// GetClientRecords returns records for a client by UUID
// @Summary Get client records
// @Description Get all records for a client by UUID
// @Tags record
// @Produce json
// @Param uuid path string true "Client UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /record/{uuid} [get]
func (h *Handler) GetClientRecords(ctx *gin.Context) {
	client_id, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		h.logger.Errorf("Handler.GetClientRecords: invalid client_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Client UUID: %v", err)})
		return
	}
	records, err := h.service.GetClientRecords(client_id)
	if err != nil {
		h.logger.Errorf("GetClientRecords: failed to get records from service layer: %v, client_id: %s", err, client_id.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered book", err)})
		return
	}

	h.logger.Infof("GetClientRecords: records retrieved successfully, client_id: %s, records_count: %d", client_id.String(), len(records))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    records,
	})
}

// GetClientRecordsFiltered returns records by user with filters
// @Summary Get filtered client records
// @Description Get client records filtered by status with pagination
// @Tags record
// @Accept json
// @Produce json
// @Param filter body object true "Filter request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /record/user/filter [post]
func (h *Handler) GetClientRecordsFiltered(ctx *gin.Context) {
	var request struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
		Page   int    `json:"page"`
		Limit  int    `json:"limit"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Handler.GetClientRecordsFiltered: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	userUUID, err := uuid.Parse(request.UserID)
	if err != nil {
		h.logger.Errorf("Handler.GetClientRecordsFiltered: bad uuid: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad user_id"})
		return
	}
	records, err := h.service.GetClientRecordsByStatus(userUUID, request.Status)
	if err != nil {
		h.logger.Errorf("Handler.GetClientRecordsFiltered: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	page := request.Page
	limit := request.Limit
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * limit
	if start > len(records) {
		start = len(records)
	}
	end := start + limit
	if end > len(records) {
		end = len(records)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Success",
		"total":    len(records),
		"page":     page,
		"limit":    limit,
		"records":  records[start:end],
		"has_next": end < len(records),
		"has_prev": start > 0,
	})
}

// GetRecordsBySlot returns records for slot by status
// @Summary Get records by slot
// @Description Get records for a specific slot with optional status filter
// @Tags record
// @Accept json
// @Produce json
// @Param request body object true "Slot and status request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /record/master/get [post]
func (h *Handler) GetRecordsBySlot(ctx *gin.Context) {
	var request struct {
		Slot_id int    `json:"slot_id"`
		Status  string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.logger.Errorf("Handler.CreateRecord: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	records, err := h.service.GetRecordsBySlot(uint(request.Slot_id), request.Status)
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered record", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"records": records,
	})
}

// GetAllRecordsBySlot returns all records for a slot (admin)
// @Summary Get records by slot (admin)
// @Description Get all records for slot by slot_id
// @Tags record
// @Produce json
// @Param slot_id path string true "Slot ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /record/master/{slot_id} [get]
func (h *Handler) GetAllRecordsBySlot(ctx *gin.Context) {
	slot_id, err := strconv.ParseUint(ctx.Param("slot_id"), 10, 0)
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: invalid slot_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Client ID: %v", err)})
		return
	}
	records, err := h.service.GetAllRecordsBySlot(uint(slot_id))
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered record", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    records,
	})
}

// ConfirmRecord confirms record
// @Summary Confirm record
// @Description Confirm record by id
// @Tags record
// @Produce json
// @Param record_id path string true "Record ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /record/master/confirm/{record_id} [post]
func (h *Handler) ConfirmRecord(ctx *gin.Context) {
	record_id, err := strconv.ParseUint(ctx.Param("record_id"), 10, 0)
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: invalid slot_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Client ID: %v", err)})
		return
	}
	err = h.service.ConfirmRecord(uint(record_id))
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered record", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// RejectRecord rejects record
// @Summary Reject record
// @Description Reject record by id
// @Tags record
// @Produce json
// @Param record_id path string true "Record ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /record/master/reject/{record_id} [post]
func (h *Handler) RejectRecord(ctx *gin.Context) {
	record_id, err := strconv.ParseUint(ctx.Param("record_id"), 10, 0)
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: invalid slot_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Client ID: %v", err)})
		return
	}
	err = h.service.RejectRecord(uint(record_id))
	if err != nil {
		h.logger.Errorf("Handler.GetRecordBySlot: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered record", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// UpdateRecordStatus updates record status
// @Summary Update record status
// @Description Update record status by id
// @Tags record
// @Accept json
// @Produce json
// @Param request body object true "Record status update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /record/master/status [post]
func (h *Handler) UpdateRecordStatus(ctx *gin.Context) {
	var req struct {
		RecordID uint   `json:"record_id"`
		Status   string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Handler.UpdateRecordStatus: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	if err := h.service.UpdateRecordStatus(req.RecordID, req.Status); err != nil {
		h.logger.Errorf("Handler.UpdateRecordStatus: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// DeleteRecord deletes a record
// @Summary Delete record
// @Description Delete record by id
// @Tags record
// @Produce json
// @Param record_id path string true "Record ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /record/master/{record_id} [delete]
func (h *Handler) DeleteRecord(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("record_id"), 10, 0)
	if err != nil {
		h.logger.Errorf("Handler.DeleteBooked: invalid id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid Client ID: %v", err)})
		return
	}
	if err := h.service.DeleteRecord(uint(id)); err != nil {
		h.logger.Errorf("Handler.DeleteBooked: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v or not registered book", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// GetUpcomingRecordsByMasterTelegramID returns upcoming confirmed records for master
// @Summary Get upcoming records for master
// @Description Get upcoming confirmed records for master by telegram_id (internal)
// @Tags record
// @Produce json
// @Param telegram_id path string true "Telegram ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /telegram/record/master/upcoming/{telegram_id} [get]
func (h *Handler) GetUpcomingRecordsByMasterTelegramID(ctx *gin.Context) {
	telegramIDStr := ctx.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.GetUpcomingRecordsByMasterTelegramID: invalid telegram_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id"})
		return
	}

	records, err := h.service.GetUpcomingRecordsByMasterTelegramID(telegramID)
	if err != nil {
		h.logger.Errorf("Handler.GetUpcomingRecordsByMasterTelegramID: service error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming records"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    records,
	})
}
