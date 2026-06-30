package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/pkg/logger"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Get().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			logger.Error("Invalid token", map[string]interface{}{"error": err})
			respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Validate token expiry
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			respondError(w, http.StatusUnauthorized, "Token expired")
			return
		}

		// Add user ID to context
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			logger.Error("Invalid user ID in token", map[string]interface{}{"error": err})
			respondError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract role from token (you'd implement this based on your token structure)
			// For now, this is a placeholder
			next.ServeHTTP(w, r)
		})
	}
}

