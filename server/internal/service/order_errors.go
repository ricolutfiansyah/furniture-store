package service

import "errors"

var (
	ErrCartEmpty               = errors.New("cart is empty")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
)
