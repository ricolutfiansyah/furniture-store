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

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unautorized")
	}

	var req service.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ShippingAddress == "" {
		response.WriteError(w, http.StatusBadRequest, "shipping address is required")
		return
	}

	resp, err := h.orderService.Checkout(r.Context(), authUser.ID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCartEmpty):
			response.WriteError(w, http.StatusBadRequest, "cart is emtpy")
		case errors.Is(err, service.ErrVariantNotFound):
			response.WriteError(w, http.StatusNotFound, "one or more variants not found")
		case errors.Is(err, service.ErrInsufficientStock):
			response.WriteError(w, http.StatusConflict, "insufficient stock")
		case errors.Is(err, service.ErrFullNameRequired):
			response.WriteError(w, http.StatusUnprocessableEntity, "full name must be filled before checkout")
		case errors.Is(err, service.ErrPhoneRequired):
			response.WriteError(w, http.StatusUnprocessableEntity, "phone number must be filled before checkout")
		default:
			log.Printf("checkout: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, resp, "order created successfully")
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unautorized")
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), authUser.ID)
	if err != nil {
		log.Printf("get orders: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, orders, "orders retrieved successfully")
}

func (h *OrderHandler) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	authUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	order, err := h.orderService.GetOrderDetail(r.Context(), authUser.ID, orderID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			response.WriteError(w, http.StatusNotFound, "order not found")
		default:
			log.Printf("get order detail: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, order, "order detail retrieved successfully")
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req service.UpdateOrderStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err = h.orderService.UpdateOrderStatus(r.Context(), orderID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			response.WriteError(w, http.StatusNotFound, "order not found")
		case errors.Is(err, service.ErrInvalidOrderStatus):
			response.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInvalidStatusTransition):
			response.WriteError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("update order status: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, nil, "order status updated successfully")
}

func (h *OrderHandler) GetOrderDetailForAdmin(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.orderService.GetOrderDetailForAdmin(r.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			response.WriteError(w, http.StatusNotFound, "order not found")
		default:
			log.Printf("get order detail for admin: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response.WriteSuccess(w, http.StatusOK, order, "order detail retrieved successfully")
}
