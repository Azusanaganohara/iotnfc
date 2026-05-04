package handlers

import (
	"iot-ktp-api/internal/services"
	"iot-ktp-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	requestingRole, _ := c.Get("user_role")
	roleStr, _ := requestingRole.(string)

	count, _ := h.svc.CountUsers()
	if count == 0 {
		roleStr = "admin"
		input.Role = "admin"
	}

	user, err := h.svc.Register(input, roleStr)
	if err != nil {
		utils.ResponseError(c, http.StatusConflict, "Registration failed", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusCreated, "User registered successfully", gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	user, tokens, err := h.svc.Login(input)
	if err != nil {
		utils.ResponseError(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Login successful", gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "refresh_token is required", err.Error())
		return
	}

	tokens, err := h.svc.RefreshToken(body.RefreshToken)
	if err != nil {
		utils.ResponseError(c, http.StatusUnauthorized, "Token refresh failed", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "Token refreshed", tokens)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.svc.GetUserByID(userID.(string))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.ResponseOK(c, http.StatusOK, "User profile", gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.ResponseOK(c, http.StatusOK, "Logged out successfully. Please discard your tokens.", nil)
}
