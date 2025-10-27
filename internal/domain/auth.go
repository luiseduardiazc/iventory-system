package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User representa un usuario del sistema
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JWTClaims representa los claims del JWT token
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest representa una petición de login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo información básica del usuario (sin password)
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// RegisterRequest representa una petición de registro
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role,omitempty"` // Optional, default: "user"
}

// Roles del sistema
const (
	RoleAdmin    = "admin"
	RoleUser     = "user"
	RoleOperator = "operator"
)

// Validate valida un usuario
func (u *User) Validate() error {
	if u.Username == "" {
		return &ValidationError{
			Field:   "username",
			Message: "username is required",
		}
	}

	if u.Email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	if u.Password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	if len(u.Password) < 6 {
		return &ValidationError{
			Field:   "password",
			Message: "password must be at least 6 characters",
		}
	}

	// Validar rol
	if u.Role != RoleAdmin && u.Role != RoleUser && u.Role != RoleOperator {
		return &ValidationError{
			Field:   "role",
			Message: "invalid role: must be admin, user, or operator",
		}
	}

	return nil
}

// ToUserInfo convierte User a UserInfo (sin password)
func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
	}
}
