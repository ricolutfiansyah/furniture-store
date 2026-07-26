package domain

import "net/http"

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound           = &AppError{Code: "user_not_found", Message: "user not found", Status: http.StatusNotFound}
	ErrEmailAlreadyRegistered = &AppError{Code: "email_already_registered", Message: "email already registered", Status: http.StatusConflict}
	ErrInvalidCredentials     = &AppError{Code: "invalid_credentials", Message: "invalid credentials", Status: http.StatusUnauthorized}

	ErrProductNotFound  = &AppError{Code: "product_not_found", Message: "product not found", Status: http.StatusNotFound}
	ErrCategoryNotFound = &AppError{Code: "category_not_found", Message: "category not found", Status: http.StatusNotFound}
	ErrVariantNotFound  = &AppError{Code: "variant_not_found", Message: "variant not found", Status: http.StatusNotFound}

	ErrCartNotFound     = &AppError{Code: "cart_not_found", Message: "cart not found", Status: http.StatusNotFound}
	ErrCartItemNotFound = &AppError{Code: "cart_item_not_found", Message: "cart item not found", Status: http.StatusNotFound}

	ErrDuplicateOrderNumber = &AppError{Code: "duplicate_order_number", Message: "order number already exists", Status: http.StatusInternalServerError}
	ErrOrderNotFound        = &AppError{Code: "order_not_found", Message: "order not found", Status: http.StatusNotFound}

	ErrAddressNotFound = &AppError{Code: "address_not_found", Message: "address not found", Status: http.StatusNotFound}

	ErrInvalidQuantity   = &AppError{Code: "invalid_quantity", Message: "quantity must be greater than 0", Status: http.StatusBadRequest}
	ErrInvalidVariantID  = &AppError{Code: "invalid_variant_id", Message: "invalid variant ID", Status: http.StatusBadRequest}
	ErrInsufficientStock = &AppError{Code: "insufficient_stock", Message: "insufficient stock", Status: http.StatusConflict}

	ErrCartEmpty = &AppError{Code: "cart_empty", Message: "cart is empty", Status: http.StatusBadRequest}

	ErrInvalidOrderStatus      = &AppError{Code: "invalid_order_status", Message: "invalid order status", Status: http.StatusBadRequest}
	ErrInvalidStatusTransition = &AppError{Code: "invalid_status_transition", Message: "invalid order status transition", Status: http.StatusConflict}

	ErrFullNameRequired = &AppError{Code: "full_name_required", Message: "full name must be filled before checkout", Status: http.StatusUnprocessableEntity}
	ErrPhoneRequired    = &AppError{Code: "phone_required", Message: "phone number must be filled before checkout", Status: http.StatusUnprocessableEntity}
)
