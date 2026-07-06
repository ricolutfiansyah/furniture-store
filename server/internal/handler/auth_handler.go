package handler

import (
	"encoding/json"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/middleware"
	"furniture-api/internal/repository"
	"furniture-api/internal/response"
	"furniture-api/internal/service"
	"furniture-api/internal/validation"
	"log"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetProfile(r.Context(), authUser.PublicID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, user, "profile retrieved successfully")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if valErrs, ok := errors.AsType[validation.ValidationErrors](err); ok {
			response.WriteValidationErrors(w, http.StatusBadRequest, valErrs)
			return
		}

		switch {
		case errors.Is(err, service.ErrEmailAlreadyRegistered):
			response.WriteError(w, http.StatusConflict, "email already registered")
		default:
			log.Printf("register error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, user, "user registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		default:
			log.Printf("login error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, resp, "login successfull")
}
