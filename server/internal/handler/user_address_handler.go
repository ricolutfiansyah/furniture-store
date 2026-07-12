package handler

import (
	"context"
	"encoding/json"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/middleware"
	"furniture-api/internal/repository"
	"furniture-api/internal/response"
	"furniture-api/internal/validation"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AddressService interface {
	CreateAddress(ctx context.Context, userID int, req domain.CreateAddressRequest) (*domain.UserAddress, error)
	UpdateAddress(ctx context.Context, id, userID int, req domain.UpdateAddressRequest) (*domain.UserAddress, error)
	DeleteAddress(ctx context.Context, id, userID int) error
	ListAddresses(ctx context.Context, userID int) ([]domain.UserAddress, error)
	SetDefaultAddress(ctx context.Context, id, userID int) error
}

type AddressHandler struct {
	addressService AddressService
}

func NewAddressHandler(s AddressService) *AddressHandler {
	return &AddressHandler{addressService: s}
}

func (h *AddressHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	address, err := h.addressService.CreateAddress(r.Context(), user.ID, req)
	if err != nil {
		if valErrs, ok := errors.AsType[validation.ValidationErrors](err); !ok {
			response.WriteValidationErrors(w, http.StatusBadRequest, valErrs)
			return
		}

		switch {
		case errors.Is(err, repository.ErrAddressNotFound):
			response.WriteError(w, http.StatusNotFound, "address not found")
		default:
			log.Printf("create address error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, address, "address created successfully")
}

func (h *AddressHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	addresses, err := h.addressService.ListAddresses(r.Context(), user.ID)
	if err != nil {
		log.Printf("list addresses error: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, addresses, "addresses retrieved successfully")
}

func (h *AddressHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid address id")
		return
	}

	var req domain.UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	address, err := h.addressService.UpdateAddress(r.Context(), id, user.ID, req)
	if err != nil {
		if valErrs, ok := errors.AsType[validation.ValidationErrors](err); ok {
			response.WriteValidationErrors(w, http.StatusBadRequest, valErrs)
			return
		}

		switch {
		case errors.Is(err, repository.ErrAddressNotFound):
			response.WriteError(w, http.StatusNotFound, "address not found")
		default:
			log.Printf("update address error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, address, "address updated successfully")
}

func (h *AddressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid address id")
		return
	}

	if err := h.addressService.DeleteAddress(r.Context(), id, user.ID); err != nil {
		switch {
		case errors.Is(err, repository.ErrAddressNotFound):
			response.WriteError(w, http.StatusNotFound, "address not found")
		default:
			log.Printf("delete address error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, nil, "address deleted successfully")
}

func (h *AddressHandler) SetDefaultAddress(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid address id")
		return
	}

	if err := h.addressService.SetDefaultAddress(r.Context(), id, user.ID); err != nil {
		switch {
		case errors.Is(err, repository.ErrAddressNotFound):
			response.WriteError(w, http.StatusNotFound, "address not found")
		default:
			log.Printf("set default address error: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, nil, "default address updated successfully")
}
