package handlers

import (
	"encoding/json"
	"net/http"

	"Extreme-Solutions/internal/api/middleware"
	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/repository"
	"github.com/google/uuid"
)

type AuthHandler struct {
	customerRepo repository.CustomerRepository
	cache        repository.CacheRepository
	cfg          *config.Config // Configuration dependency property tracking field
}

// NewAuthHandler incorporates the config dependency pointer
func NewAuthHandler(customerRepo repository.CustomerRepository, cache repository.CacheRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		customerRepo: customerRepo,
		cache:        cache,
		cfg:          cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement authentication credentials database validation loop.

	// Mock an active user profile entity target for sandbox route testing
	mockUserID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	mockEmail := "admin@extreme-solutions.com"
	mockRole := "admin"

	// Generate a authentic signed JWT session string leveraging our middleware utility
	token, err := middleware.GenerateToken(
		h.cfg.JWT.Secret, // Pulled from config.yaml
		mockUserID,
		mockEmail,
		mockRole,
		h.cfg.JWT.AccessDuration, // Automatically scales using your 15m duration specification
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create runtime auth token context"})
		return
	}

	// Format response to provide the security string payload back to terminal
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful ready",
		"token":   token,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement customer registration flow
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Registration successful"}`))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement JWT token token lifecycle refresh rotation
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"token": "new_refreshed_jwt_token_string"}`))
}
