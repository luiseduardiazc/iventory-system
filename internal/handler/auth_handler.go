package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"inventory-system/internal/domain"
	"inventory-system/internal/service"
)

// AuthHandler maneja las peticiones HTTP de autenticación
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler crea una nueva instancia del handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Registrar un nuevo usuario
// @Description Crea un nuevo usuario en el sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Datos del usuario"
// @Success 201 {object} domain.User
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	// No retornar password
	c.JSON(http.StatusCreated, user.ToUserInfo())
}

// Login godoc
// @Summary Login de usuario
// @Description Autentica un usuario y retorna un JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Credenciales"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	loginResp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

// RefreshToken godoc
// @Summary Refrescar token
// @Description Genera un nuevo token a partir de uno válido
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Obtener token del header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "missing authorization header",
		})
		return
	}

	// Remover "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	loginResp, err := h.authService.RefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

// GetProfile godoc
// @Summary Obtener perfil del usuario
// @Description Retorna la información del usuario autenticado
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserInfo
// @Failure 401 {object} ErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Obtener user ID del contexto (lo pone el middleware de auth)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	user, err := h.authService.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user.ToUserInfo())
}
