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
)

type AddressService interface {
	CreateAddress(ctx context.Context, userID int, req domain.CreateAddressRequest) (*domain.UserAddress, error)
	UpdateAddress(ctx context.Context, id, userID int, req domain.UpdateAddressRequest) (*domain.UserAddress, error)
	DeleteAddress(ctx context.Context, id, userID int) error
	ListAddresses(ctx context.Context, userID int) ([]domain.UserAddress, error)
	SetDefaultAddress(ctx context.Context, id, userID int)
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
