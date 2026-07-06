package handler

import (
	"encoding/json"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/middleware"
	"furniture-api/internal/response"
	"furniture-api/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
	}

	var req domain.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid body request")
		return
	}

	if req.VariantID <= 0 {
		response.WriteError(w, http.StatusBadRequest, "variant id is required")
	}

	err := h.cartService.AddToCart(r.Context(), authUser.ID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuantity):
			response.WriteError(w, http.StatusBadRequest, "quantity must be greater than 0")
		case errors.Is(err, service.ErrVariantNotFound):
			response.WriteError(w, http.StatusNotFound, "variant not found")
		case errors.Is(err, service.ErrInsufficientStock):
			response.WriteError(w, http.StatusConflict, "insufficient stock")
		default:
			log.Printf("add to cart: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, nil, "item added to cart successfully")
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart, err := h.cartService.GetCart(r.Context(), authUser.ID)
	if err != nil {
		log.Printf("get cart: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, cart, "cart retrieved successfully")
}

func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err = h.cartService.UpdateQuantity(r.Context(), authUser.ID, itemID, req.Quantity); err != nil {
		switch {
		case errors.Is(err, service.ErrCartItemNotFound):
			response.WriteError(w, http.StatusNotFound, "cart item not found")
		default:
			log.Printf("update quantity: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, nil, "quantity updated successfully")
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err = h.cartService.RemoveItem(r.Context(), authUser.ID, itemID); err != nil {
		switch {
		case errors.Is(err, service.ErrCartItemNotFound):
			response.WriteError(w, http.StatusNotFound, "cart item not found")
		default:
			log.Printf("remove item: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, nil, "item removed from cart")
}
