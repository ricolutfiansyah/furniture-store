package service

import (
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
)

func toNullString(s string) nullable.NullString {
	if s == "" {
		return nullable.NewNull()
	}

	return nullable.NewNullString(s)
}

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address"`
	Notes           string `json:"notes"`
}

type CheckoutResponse struct {
	Order      domain.Order       `json:"order"`
	Items      []domain.OrderItem `json:"items"`
	GrandTotal float64            `json:"grand_total"`
}

type UpdateOrderStatusReq struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}
