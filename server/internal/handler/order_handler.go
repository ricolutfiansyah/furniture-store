package handler

import (
	"encoding/json"
	"furniture-api/internal/middleware"
	"furniture-api/internal/service"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	var req service.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShippingAddress == "" {
		http.Error(w, "Shipping address is required", http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.Checkout(r.Context(), userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resp,
		"message": "Order created successfully",
	})
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    orders,
		"message": "Orders retrieved successfully",
	})
}

func (h *OrderHandler) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	order, err := h.orderService.GetOrderDetail(r.Context(), userID, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    order,
		"message": "Order detail retrieved successfully",
	})
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req service.UpdateOrderStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adminName := "admin"

	err = h.orderService.UpdateOrderStatus(r.Context(), orderID, req, adminName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Order status updated successfully",
	})
}
