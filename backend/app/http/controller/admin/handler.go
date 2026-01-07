package admin

import (
	"app/encoder"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminLogin handles admin authentication request
// @Summary Admin login
// @Description Authenticate admin user with password
// @Tags admin
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin login credentials"
// @Success 200 {object} AdminLoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/login [post]
func (h *Handler) AdminLogin(ctx *gin.Context) {
	var req AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("AdminLogin: invalid request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	expectedPassword := os.Getenv("ADMIN_PASSWORD")
	if expectedPassword == "" {
		h.logger.Errorf("AdminLogin: ADMIN_PASSWORD environment variable is not set")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}
	if req.Password != expectedPassword {
		h.logger.Errorf("AdminLogin: invalid password attempt")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Generate JWT token for admin (using fixed ID)
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	token, err := encoder.GenerateToken(adminID)
	if err != nil {
		h.logger.Errorf("Handler.AdminLogin: failed to generate token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Admin login successful",
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":         adminID.String(),
			"first_name": "Admin",
			"surname":    "System",
			"phone":      "admin",
		},
	})
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete user by ID
// @Tags admin
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/user/{id} [delete]
func (h *Handler) DeleteUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.logger.Errorf("DeleteUser: invalid user ID format: %v, param: %s", err, ctx.Param("id"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userRepo.DeleteUser(userID)
	if err != nil {
		h.logger.Errorf("DeleteUser: failed to delete user from database: %v, user_id: %s", err, userID.String())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	h.logger.Infof("DeleteUser: user deleted successfully, user_id: %s", userID.String())

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user_id": userID.String(),
	})
}

// ToggleUserActive toggles user active status
// @Summary Toggle user active status
// @Description Toggle user active/inactive status
// @Tags admin
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/user/{id}/toggle-active [put]
func (h *Handler) ToggleUserActive(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.logger.Errorf("ToggleUserActive: invalid user ID format: %v, param: %s", err, ctx.Param("id"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		h.logger.Errorf("ToggleUserActive: failed to get user from database: %v, user_id: %s", err, userID.String())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	newActive := !user.Active
	err = h.userRepo.UpdateUserActive(userID, newActive)
	if err != nil {
		h.logger.Errorf("ToggleUserActive: failed to update user active status: %v, user_id: %s, new_active: %v", err, userID.String(), newActive)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	h.logger.Infof("ToggleUserActive: user active status updated, user_id: %s, old_active: %v, new_active: %v", userID.String(), user.Active, newActive)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
		"user_id": userID.String(),
		"active":  newActive,
	})
}

func (h *Handler) GetSlots(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: invalid master_id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	slotsWithMaster, err := h.slotRepo.FindSlots(userID)
	if err != nil {
		h.logger.Errorf("Handler.GetSlots: service error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	h.logger.Infof("Handler.GetSlots: sending slots for master_id=%v", userID)
	ctx.JSON(http.StatusOK, slotsWithMaster)
}

// DeleteSlot deletes a slot
// @Summary Delete slot
// @Description Delete slot by ID
// @Tags admin
// @Produce json
// @Param id path string true "Slot ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/slot/{id} [delete]
func (h *Handler) DeleteSlot(ctx *gin.Context) {
	slotIDStr := ctx.Param("id")
	slotID, err := strconv.ParseUint(slotIDStr, 10, 32)
	if err != nil {
		h.logger.Errorf("DeleteSlot: invalid slot ID format: %v, param: %s", err, slotIDStr)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	err = h.slotRepo.DeleteSlot(uint(slotID))
	if err != nil {
		h.logger.Errorf("DeleteSlot: failed to delete slot from database: %v, slot_id: %d", err, slotID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete slot"})
		return
	}

	h.logger.Infof("DeleteSlot: slot deleted successfully, slot_id: %d", slotID)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Slot deleted successfully",
		"slot_id": slotID,
	})
}
