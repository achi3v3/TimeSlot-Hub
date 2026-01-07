package user

import (
	"app/http/sender"
	ucase "app/http/usecase/user"
	"app/pkg/models"
	"fmt"
	_ "fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserSageRegister struct {
	Phone      string `json:"phone"       gorm:"unique; not null; column:phone"`
	TelegramID int64  `json:"telegram_id" gorm:"index; column:telegram_id"`
	FirstName  string `json:"first_name"  gorm:"column:first_name; not null"`
	Surname    string `json:"surname"     gorm:"column:surname"`
}

// CreateUser creates a new user
// @Summary Create user
// @Description Register a new user with phone and Telegram ID
// @Tags user
// @Accept json
// @Produce json
// @Param user body UserSageRegister true "User registration data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /user/register [post]
func (h *Handler) CreateUser(ctx *gin.Context) {
	var safeUser UserSageRegister
	if err := ctx.ShouldBindJSON(&safeUser); err != nil {
		h.logger.Errorf("Handler.CreateUser: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request")})
		return
	}

	// Validate input data
	if len(safeUser.Phone) < 10 || len(safeUser.Phone) > 15 {
		h.logger.Errorf("CreateUser: invalid phone number length: %s", safeUser.Phone)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		return
	}

	if len(safeUser.FirstName) < 1 || len(safeUser.FirstName) > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "First name must be between 1 and 50 characters"})
		return
	}

	if len(safeUser.Surname) > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Surname must be less than 50 characters"})
		return
	}

	if safeUser.TelegramID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}

	user := models.User{
		Phone:      safeUser.Phone,
		TelegramID: safeUser.TelegramID,
		FirstName:  safeUser.FirstName,
		Surname:    safeUser.Surname,
	}
	if err := h.service.Register(&user); err != nil {
		h.logger.Errorf("CreateUser: failed to register user in service layer: %v, phone: %s, telegram_id: %d", err, safeUser.Phone, safeUser.TelegramID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Register error"})
		return
	}

	h.logger.Infof("CreateUser: user registered successfully, phone: %s, telegram_id: %d", safeUser.Phone, safeUser.TelegramID)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User created successfully. Confirm via Telegram-Bot",
		"result":  true,
	})
}

// Login initiates user login process
// @Summary User login
// @Description Initiate login process by phone number
// @Tags user
// @Accept json
// @Produce json
// @Param request body object true "Login request" example({"phone": "+79991234567"})
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /user/login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var jsonData struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&jsonData); err != nil {
		h.logger.Errorf("Handler.Login: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate phone number format
	if len(jsonData.Phone) < 10 || len(jsonData.Phone) > 15 {
		h.logger.Errorf("Login: invalid phone number format: %s", jsonData.Phone)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		return
	}

	user, message, err := h.service.Login(jsonData.Phone)
	if err != nil {
		h.logger.Errorf("Login: failed to process login in service layer: %v, phone: %s", err, jsonData.Phone)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Login error"})
		return
	}

	// Send notification to Telegram with IP/location (best-effort)
	if user != nil {
		ip := ctx.ClientIP()
		city := ctx.GetHeader("X-Geo-City")
		country := ctx.GetHeader("X-Geo-Country")
		loc := ""
		if city != "" && country != "" {
			loc = city + "," + country
		} else if country != "" {
			loc = country
		}
		if err := sender.LoginNotify(*user, ip, loc); err != nil {
			h.logger.Errorf("Login: failed to send Telegram notification: %v, user_id: %s", err, user.ID.String())
		} else {
			h.logger.Infof("Login: Telegram notification sent successfully, user_id: %s, ip: %s, location: %s", user.ID.String(), ip, loc)
		}
	}

	// Set short-lived cookie to allow claim-token from frontend without secret
	// Cookie expires in ~3 minutes
	ctx.SetCookie("login_flow", "1", 180, "/", "", false, true)

	h.logger.Infof("Login: login request processed successfully, phone: %s", jsonData.Phone)
	var resp gin.H = gin.H{"message": message}
	if user != nil {
		resp["user"] = gin.H{
			"id":          user.ID,
			"first_name":  user.FirstName,
			"surname":     user.Surname,
			"telegram_id": user.TelegramID,
			"phone":       user.Phone,
			"roles":       user.Roles,
			"timezone":    user.Timezone,
		}
	}
	ctx.JSON(http.StatusOK, resp)
}

// CheckAuth checks if user is authenticated by telegram_id
// @Summary Check auth
// @Description Check authentication status by telegram_id
// @Tags user
// @Produce json
// @Param telegram_id path int true "Telegram ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /user/check/{telegram_id} [get]
func (h *Handler) CheckAuth(ctx *gin.Context) {
	telegramID, err := strconv.ParseInt(ctx.Param("telegram_id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.CheckAuth: invalid telegram id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}

	user, err := h.service.GetByTelegramID(telegramID)
	if err != nil {
		h.logger.Errorf("CheckAuth: failed to get user from service layer: %v, telegram_id: %d", err, telegramID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User inactive"})
		return
	}
	if user == nil {
		h.logger.Infof("CheckAuth: user not found, telegram_id: %d", telegramID)
		ctx.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}

	h.logger.Infof("CheckAuth: user authentication check completed, telegram_id: %d, user_id: %s, active: %v", telegramID, user.ID.String(), user.Active)
	ctx.JSON(http.StatusOK, gin.H{"authenticated": true})
}

type PublicUserResponse struct {
	ID        string           `json:"id"`
	FirstName string           `json:"first_name"`
	Surname   string           `json:"surname"`
	Services  []models.Service `json:"services,omitempty"`
}

// GetUserByTelegramID returns user public profile by telegram_id (internal)
// @Summary Get user by telegram id (internal)
// @Description Get user public data by telegram id (internal)
// @Tags user
// @Produce json
// @Param telegram_id path int true "Telegram ID"
// @Success 200 {object} PublicUserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user/g3tter/{telegram_id} [get]
func (h *Handler) GetUserByTelegramID(ctx *gin.Context) {
	telegramID, err := strconv.ParseInt(ctx.Param("telegram_id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.GetUserByTelegramID: invalid telegram id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}

	user, err := h.service.GetByTelegramID(telegramID)
	if err != nil {
		h.logger.Errorf("Handler.GetUserByTelegramID: user not registered or inactive: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User inactive or not registered"})
		return
	}
	publicUser := PublicUserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		Surname:   user.Surname,
		Services:  user.Services,
	}

	ctx.JSON(http.StatusOK, publicUser)
}

// ConfirmLogin confirms login by telegram_id (internal)
// @Summary Confirm login
// @Description Confirm login by telegram_id (internal endpoint)
// @Tags user
// @Produce json
// @Param telegram_id path int true "Telegram ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user/confirm-login/{telegram_id} [post]
func (h *Handler) ConfirmLogin(ctx *gin.Context) {
	telegramID, err := strconv.ParseInt(ctx.Param("telegram_id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.ConfirmLogin: invalid telegram id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}
	if err := h.service.ConfirmLoginByTelegramID(telegramID); err != nil {
		h.logger.Errorf("Handler.ConfirmLogin: user not found: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.logger.Infof("Handler.ConfirmLogin: token issued for telegram_id=%d", telegramID)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Token issued",
	})
}

// ClaimToken issues session token for user (internal)
// @Summary Claim token
// @Description Issue session token for user after confirmation (internal)
// @Tags user
// @Produce json
// @Param telegram_id path int true "Telegram ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user/claim-token/{telegram_id} [post]
func (h *Handler) ClaimToken(ctx *gin.Context) {
	telegramID, err := strconv.ParseInt(ctx.Param("telegram_id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.ClaimToken: invalid telegram id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}
	// Retrieve temporary token from memory (signal from confirmation)
	claimedToken, err := h.service.ClaimUserTokenByTelegramID(telegramID)
	if err != nil {
		h.logger.Errorf("ClaimToken: user not found: %v, telegram_id: %d", err, telegramID)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// If temporary token is not ready, return pending status
	if claimedToken == "" {
		h.logger.Infof("ClaimToken: token not ready yet, telegram_id: %d", telegramID)
		ctx.JSON(http.StatusOK, gin.H{"message": "Token not ready", "pending": true})
		return
	}

	// Prepare user data for response (without JWT)
	user, _ := h.service.GetByTelegramID(telegramID)
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.logger.Infof("ClaimToken: token issued successfully, telegram_id: %d", telegramID)
	// Return session temporary token without JWT, including additional fields
	safeUser := gin.H{
		"id":          user.ID.String(),
		"first_name":  user.FirstName,
		"surname":     user.Surname,
		"telegram_id": user.TelegramID,
		"phone":       user.Phone,
		"roles":       user.Roles,
		"timezone":    user.Timezone,
	}
	// Clear login_flow cookie on successful token issuance
	ctx.SetCookie("login_flow", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Token issued",
		"token":   claimedToken,
		"user":    safeUser,
	})
}

// CheckLogin checks pending login token status
// @Summary Check login status
// @Description Check if login token is ready for given telegram_id
// @Tags user
// @Produce json
// @Param telegram_id path int true "Telegram ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /user/check-login/{telegram_id} [get]
func (h *Handler) CheckLogin(ctx *gin.Context) {
	telegramID, err := strconv.ParseInt(ctx.Param("telegram_id"), 10, 64)
	if err != nil {
		h.logger.Errorf("Handler.CheckLogin: invalid telegram id: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Telegram ID"})
		return
	}
	pending := h.service.CheckUserTokenByTelegramID(telegramID)

	h.logger.Infof("Handler.CheckLogin: send pending %t", pending)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Token in storage",
		"pending": pending,
	})
}

// Logout user session
// @Summary Logout
// @Description Logout current session
// @Tags user
// @Produce json
// @Success 200 {object} map[string]string
// @Router /user/logout [post]
func (h *Handler) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetPublic returns public user info by UUID
// @Summary Get public user
// @Description Get public user info by UUID
// @Tags user
// @Produce json
// @Param uuid path string true "User UUID"
// @Success 200 {object} PublicUserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /user/public/{uuid} [get]
func (h *Handler) GetPublic(ctx *gin.Context) {
	idStr := ctx.Param("uuid")
	user, err := h.service.GetPublicByID(idStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}
	safeUser := PublicUserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		Surname:   user.Surname,
		Services:  user.Services,
	}
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	ctx.JSON(http.StatusOK, safeUser)
}

// UpdateUser updates basic user info
// @Summary Update user names
// @Description Update first_name and surname for user
// @Tags user
// @Accept json
// @Produce json
// @Param request body map[string]string true "User update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/update [put]
func (h *Handler) UpdateUser(ctx *gin.Context) {
	var body struct {
		FirstName string `json:"first_name"`
		Surname   string `json:"surname"`
		UserID    string `json:"user_id"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Errorf("Handler.UpdateUser: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate input data
	if len(body.FirstName) < 1 || len(body.FirstName) > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "First name must be between 1 and 50 characters"})
		return
	}

	if len(body.Surname) > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Surname must be less than 50 characters"})
		return
	}

	req := ucase.UpdateNamesRequest{
		UserID:    body.UserID,
		FirstName: body.FirstName,
		Surname:   body.Surname,
	}

	// use service method
	if err := h.service.UpdateNames(req); err != nil {
		h.logger.Errorf("Handler.UpdateUser: update error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// UpdateTimezone sets user timezone
// @Summary Update timezone
// @Description Update timezone for authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Param request body map[string]string true "Timezone update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/timezone [put]
func (h *Handler) UpdateTimezone(ctx *gin.Context) {
	var body struct {
		UserID   string `json:"user_id"`
		Timezone string `json:"timezone"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Errorf("Handler.UpdateTimezone: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	req := ucase.UpdateTimezoneRequest{UserID: body.UserID, Timezone: body.Timezone}
	if err := h.service.UpdateTimezone(req); err != nil {
		h.logger.Errorf("Handler.UpdateTimezone: update error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Timezone updated successfully"})
}

// UpdateTimezoneInternal updates timezone by telegram_id (internal, Telegram)
// @Summary Update timezone internal
// @Description Update timezone by telegram_id (internal for Telegram bot)
// @Tags user
// @Accept json
// @Produce json
// @Param request body map[string]string true "Timezone update internal request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /telegram/user/timezone [put]
func (h *Handler) UpdateTimezoneInternal(ctx *gin.Context) {
	var body struct {
		TelegramID int64  `json:"telegram_id"`
		Timezone   string `json:"timezone"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		h.logger.Errorf("Handler.UpdateTimezoneInternal: invalid request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if body.TelegramID == 0 || body.Timezone == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "telegram_id and timezone are required"})
		return
	}
	user, err := h.service.GetByTelegramID(body.TelegramID)
	if err != nil || user == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	req := ucase.UpdateTimezoneRequest{UserID: user.ID.String(), Timezone: body.Timezone}
	if err := h.service.UpdateTimezone(req); err != nil {
		h.logger.Errorf("Handler.UpdateTimezoneInternal: update error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Timezone updated successfully"})
}

func (h *Handler) GetPublicUser(ctx *gin.Context) {
	userID := ctx.Param("uuid")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	user, err := h.service.GetPublicByID(userID)
	if err != nil {
		h.logger.Errorf("Handler.GetPublicUser: error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusOK, gin.H{"user": nil})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": user.ID, "first_name": user.FirstName, "surname": user.Surname})
}

// RequestAccountDeletion requests account deletion via Telegram
// @Summary Request account deletion
// @Description Request account deletion, sending confirmation to Telegram
// @Tags user
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/request-deletion [post]
func (h *Handler) RequestAccountDeletion(ctx *gin.Context) {
	userUUIDInterface, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("Handler.RequestAccountDeletion: user_id not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userUUIDInterface.(uuid.UUID)
	if !ok {
		h.logger.Errorf("Handler.RequestAccountDeletion: invalid user_id type in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user to verify telegram_id
	user, err := h.service.GetUserByID(userUUID)
	if err != nil {
		h.logger.Errorf("RequestAccountDeletion: failed to get user: %v, user_id: %s", err, userUUID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if user.TelegramID == 0 {
		h.logger.Errorf("RequestAccountDeletion: telegram_id not found, user_id: %s", userUUID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Telegram ID not found. Cannot send confirmation"})
		return
	}

	// Send confirmation request to Telegram
	err = h.service.RequestAccountDeletion(userUUID, user.TelegramID)
	if err != nil {
		h.logger.Errorf("Handler.RequestAccountDeletion: failed to send telegram confirmation: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Confirmation request sent to Telegram",
	})
}

// ConfirmAccountDeletion confirms account deletion
// @Summary Confirm account deletion
// @Description Confirm account deletion (internal for Telegram)
// @Tags user
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /user/confirm-deletion [post]
func (h *Handler) ConfirmAccountDeletion(ctx *gin.Context) {
	// Get user_id from header (for internal calls from Telegram bot)
	userIDStr := ctx.GetHeader("X-User-ID")
	if userIDStr == "" {
		// If not in header, try to get from context (for regular requests)
		userUUIDInterface, exists := ctx.Get("user_id")
		if !exists {
			h.logger.Errorf("ConfirmAccountDeletion: user_id not found in context or header")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		userUUID, ok := userUUIDInterface.(uuid.UUID)
		if !ok {
			h.logger.Errorf("ConfirmAccountDeletion: invalid user_id type in context")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		// Delete user and all related data
		err := h.service.DeleteUser(userUUID)
		if err != nil {
			h.logger.Errorf("ConfirmAccountDeletion: failed to delete user: %v, user_id: %s", err, userUUID.String())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
			return
		}
		h.logger.Infof("ConfirmAccountDeletion: account deleted successfully, user_id: %s", userUUID.String())
	} else {
		// Parse user_id from header
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Errorf("ConfirmAccountDeletion: invalid user_id in header: %v, header_value: %s", err, userIDStr)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
		// Delete user and all related data
		err = h.service.DeleteUser(userUUID)
		if err != nil {
			h.logger.Errorf("ConfirmAccountDeletion: failed to delete user: %v, user_id: %s", err, userUUID.String())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
			return
		}
		h.logger.Infof("ConfirmAccountDeletion: account deleted successfully, user_id: %s", userUUID.String())
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Account successfully deleted",
	})
}
