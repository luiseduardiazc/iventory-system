package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

// AuthService maneja la lógica de autenticación
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

// NewAuthService crea una nueva instancia del servicio
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

// Register registra un nuevo usuario
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Validar que no exista el username
	existingByUsername, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingByUsername != nil {
		return nil, &domain.ConflictError{
			Message: fmt.Sprintf("username %s already exists", req.Username),
		}
	}

	// Validar que no exista el email
	existingByEmail, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingByEmail != nil {
		return nil, &domain.ConflictError{
			Message: fmt.Sprintf("email %s already exists", req.Email),
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determinar rol (default: user)
	role := req.Role
	if role == "" {
		role = domain.RoleUser
	}

	// Crear usuario
	user := &domain.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      role,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validar
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Guardar en BD
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login autentica un usuario y retorna un JWT token
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Buscar usuario
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, &domain.UnauthorizedError{
			Message: "invalid credentials",
		}
	}

	// Verificar que esté activo
	if !user.Active {
		return nil, &domain.UnauthorizedError{
			Message: "user is not active",
		}
	}

	// Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, &domain.UnauthorizedError{
			Message: "invalid credentials",
		}
	}

	// Generar JWT token
	token, expiresAt, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToUserInfo(),
	}, nil
}

// GenerateToken genera un JWT token para el usuario
func (s *AuthService) GenerateToken(user *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)

	claims := &domain.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "inventory-system",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken valida un JWT token y retorna los claims
func (s *AuthService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, &domain.UnauthorizedError{
			Message: "invalid token",
		}
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, &domain.UnauthorizedError{
			Message: "invalid token claims",
		}
	}

	return claims, nil
}

// GetUser obtiene un usuario por ID
func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// RefreshToken genera un nuevo token a partir de uno válido
func (s *AuthService) RefreshToken(ctx context.Context, oldTokenString string) (*domain.LoginResponse, error) {
	// Validar token actual
	claims, err := s.ValidateToken(oldTokenString)
	if err != nil {
		return nil, err
	}

	// Obtener usuario actualizado
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.Active {
		return nil, &domain.UnauthorizedError{
			Message: "user is not active",
		}
	}

	// Generar nuevo token
	token, expiresAt, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToUserInfo(),
	}, nil
}
