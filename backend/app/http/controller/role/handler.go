package role

import (
	"app/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRole creates a new role for a user
// @Summary Create role
// @Description Create role for user
// @Tags role
// @Accept json
// @Produce json
// @Param role body models.UserRole true "Role data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var req models.UserRole
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.roleService.CreateRole(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"user_id": req.UserID,
		"role":    req.Role,
	})
}

// DeleteRole removes a role from a user
// @Summary Delete role
// @Description Delete role from user
// @Tags role
// @Accept json
// @Produce json
// @Param role body models.UserRole true "Role data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	var req models.UserRole
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.roleService.DeleteRole(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
		"user_id": req.UserID,
		"role":    req.Role,
	})
}

// GetUserRoles retrieves all roles for a specific user
// @Summary Get user roles
// @Description Get roles for user by ID
// @Tags role
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{id}/roles [get]
func (h *Handler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	response, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAllRoles retrieves all roles in the system
// @Summary Get all roles
// @Description Get all roles
// @Tags role
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /admin/roles [get]
func (h *Handler) GetAllRoles(c *gin.Context) {
	response, err := h.roleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CheckUserRole checks if a user has a specific role
// @Summary Check user role
// @Description Check if user has specific role
// @Tags role
// @Produce json
// @Param id path string true "User ID"
// @Param role path string true "Role name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{id}/roles/{role} [get]
func (h *Handler) CheckUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleName := c.Param("role")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role name is required"})
		return
	}

	exists, err := h.roleService.CheckUserRole(userID, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"role":    roleName,
		"exists":  exists,
	})
}
