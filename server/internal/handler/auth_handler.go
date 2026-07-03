package handler

import (
	"encoding/json"
	"errors"
	"furniture-api/internal/middleware"
	"furniture-api/internal/repository"
	"furniture-api/internal/service"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	publicID, ok := claims["sub"].(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	user, err := h.authService.GetProfile(r.Context(), publicID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeSuccess(w, http.StatusOK, user, "profile retrieved successfully")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    user,
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    resp,
		"message": "Login successfull",
	})
}
