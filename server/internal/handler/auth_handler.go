package handler

import (
	"context"
	"encoding/json"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/middleware"
	"furniture-api/internal/response"
	"furniture-api/internal/validation"
	"log"
	"net/http"
)

type AuthService interface {
	GetProfile(ctx context.Context, publicID string) (*domain.UserResponse, error)
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
}

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(s AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authService.GetProfile(r.Context(), authUser.PublicID)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			response.WriteError(w, appErr.Status, appErr.Message)
			return
		}
		log.Printf("get profile error: %v", err)
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

		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			response.WriteError(w, appErr.Status, appErr.Message)
			return
		}

		log.Printf("register error: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusCreated, user, "user registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if valErrs, ok := errors.AsType[validation.ValidationErrors](err); ok {
			response.WriteValidationErrors(w, http.StatusBadRequest, valErrs)
			return
		}

		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			response.WriteError(w, appErr.Status, appErr.Message)
			return
		}

		log.Printf("login error: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, resp, "login successful")
}
