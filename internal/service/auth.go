package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/domain/repository"
	"github.com/jmoloko/taskmange/internal/logger"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("invalid password: must be at least 6 characters")
)

// Сервис аутентификации
type AuthService struct {
	repo   repository.UserRepository
	logger logger.Logger
	secret string
}

func NewAuthService(repo repository.UserRepository, logger logger.Logger, secret string) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
		secret: secret,
	}
}

// регистрация нового пользователя
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) error {
	// валидация email
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return ErrInvalidEmail
	}

	// валидация пароля
	if len(req.Password) < 6 {
		return ErrInvalidPassword
	}

	// проверка на существование пользователя в базе
	existingUser, _ := s.repo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return ErrUserExists
	}

	// хэшируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           generateUUID(),
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.repo.Create(ctx, user)
}

// аутентификация пользователя и возврат токена
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	// Find user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// создание токена
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// валидируем токен и возвращаем Id пользователя
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", ErrInvalidToken
		}
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// получаем пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// генерация токена
func (s *AuthService) generateToken(userID string) (string, error) {
	// Create token claims
	expirationTime := time.Now().Add(time.Minute * 15)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateUUID() string {
	return uuid.New().String()
}
