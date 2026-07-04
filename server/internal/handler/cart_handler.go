package handler

import (
	"encoding/json"
	"errors"
	"furniture-api/internal/middleware"
	"furniture-api/internal/response"
	"furniture-api/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
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

	var req service.AddToCartRequest
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
			log.Printf("add to cart: %w", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, nil, "item added to cart successfully")
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	cart, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    cart,
		"message": "Cart retrieved successfully",
	})
}

func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.cartService.UpdateQuantity(r.Context(), userID, itemID, req.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Quantity updated successfully",
	})
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	itemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	err = h.cartService.RemoveItem(r.Context(), userID, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Item removed from cart",
	})
}
