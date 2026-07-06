package service

import "errors"

var (
	ErrCartEmpty               = errors.New("cart is empty")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
	ErrFullNameRequired        = errors.New("full name must be filled before checkout")
	ErrPhoneRequired           = errors.New("phone number must be filled before checkout")
	ErrShippingAddressEmpty    = errors.New("Shipping address is required")
)
