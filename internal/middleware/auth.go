package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/logger"
)

// userIDKey — это ключ для хранения идентификатора пользователя в контексте
type userIDKey struct{}

// AuthService интерфейс для аутентификации
type AuthService interface {
	ValidateToken(token string) (string, error)
}

// AuthMiddleware проверка JWT
func AuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// проверка формата токена
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected 'Bearer <token>'"})
			c.Abort()
			return
		}

		token := parts[1]

		// валидация токена
		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// добавление ID user в контекст
		c.Set("user_id", userID)
		c.Next()
	}
}

// Auth middleware for JWT token validation
func Auth(authService AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			tokenString, err := extractTokenFromHeader(r)
			if err != nil {
				logger.Error("Authentication failed: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token
			userID, err := authService.ValidateToken(tokenString)
			if err != nil {
				logger.Error("Token validation failed: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

// extractTokenFromHeader extracts JWT token from Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
