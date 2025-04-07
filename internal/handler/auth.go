package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/service"
)

// AuthHandler handles authentication HTTP requests using Gin
type AuthHandler struct {
	service *service.AuthService
	logger  logger.Logger
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(service *service.AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

// Register регистрация пользователя
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} map[string]interface{} "Registration successful"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 409 {object} map[string]string "Conflict - User already exists"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode register request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.Register(c.Request.Context(), req); err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		case service.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		case service.ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		default:
			h.logger.Error("Failed to register user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
	})
}

// Login аутентификация пользователя
// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} map[string]string "Token"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid credentials"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		h.logger.Error("Failed to login user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetService возвращает сервис аутентификации
func (h *AuthHandler) GetService() *service.AuthService {
	return h.service
}
