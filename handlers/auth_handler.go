package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register request"
// @Success 201 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	if err := h.authService.Register(&req); err != nil {
		if err.Error() == "user with email "+req.Email+" already exists" {
			return utils.Conflict(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to register user")
	}

	return utils.SuccessResponse(c, http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return utils.Unauthorized(c, "Invalid email or password")
		}
		return utils.InternalServerError(c, "Failed to login")
	}

	return utils.SuccessResponse(c, http.StatusOK, response)
}

// GetMe godoc
// @Summary Get current user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Router /api/auth/me [get]
func (h *AuthHandler) GetMe(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return utils.Unauthorized(c, "Unauthorized")
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		return utils.NotFound(c, "User not found")
	}

	return utils.SuccessResponse(c, http.StatusOK, user)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshRequest true "Refresh request"
// @Success 200 {object} models.LoginResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req models.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	response, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return utils.Unauthorized(c, "Invalid or expired refresh token")
	}

	return utils.SuccessResponse(c, http.StatusOK, response)
}

// Logout godoc
// @Summary Logout user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshRequest true "Logout request"
// @Success 200 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	var req models.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		return utils.InternalServerError(c, "Failed to logout")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
