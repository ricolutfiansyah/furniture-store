package service

import "errors"

var (
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrVariantNotFound   = errors.New("variant not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrCartItemNotFound  = errors.New("cart item not found")
)
