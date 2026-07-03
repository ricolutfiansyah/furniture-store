package repository

import "errors"

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrVariantNotFound   = errors.New("variant not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)
