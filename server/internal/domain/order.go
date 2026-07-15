package domain

import (
	"furniture-api/internal/nullable"
	"time"
)

type Order struct {
	ID              int                 `db:"id" json:"id"`
	UserID          int                 `db:"user_id" json:"user_id"`
	OrderNumber     string              `db:"order_number" json:"order_number"`
	TotalAmount     float64             `db:"total_amount" json:"total_amount"`
	ShippingCost    float64             `db:"shipping_cost" json:"shipping_cost"`
	Tax             float64             `db:"tax" json:"tax"`
	GrandTotal      float64             `db:"grand_total" json:"grand_total"`
	Status          string              `db:"status" json:"status"`
	ShippingAddress string              `db:"shipping_address" json:"shipping_address"`
	PaymentMethod   string              `db:"payment_method" json:"payment_method"`
	PaidAt          nullable.NullTime   `db:"paid_at" json:"paid_at"`
	ShippedAt       nullable.NullTime   `db:"shipped_at" json:"shipped_at"`
	DeliveredAt     nullable.NullTime   `db:"delivered_at" json:"delivered_at"`
	Notes           nullable.NullString `db:"notes" json:"notes"`
	CreatedAt       time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `db:"updated_at" json:"updated_at"`

	Items    []OrderItem   `db:"-" json:"items,omitempty"`
	Statuses []OrderStatus `db:"-" json:"statuses,omitempty"`
	User     *User         `db:"-" json:"user,omitempty"`

	FirstItemName  string `db:"-" json:"first_item_name,omitempty"`
	FirstItemImage string `db:"-" json:"first_item_image,omitempty"`
	TotalItems     int    `db:"-" json:"total_items,omitempty"`
}

type OrderItem struct {
	ID           int       `db:"id" json:"id"`
	OrderID      int       `db:"order_id" json:"order_id"`
	VariantID    int       `db:"variant_id" json:"variant_id"`
	Quantity     int       `db:"quantity" json:"quantity"`
	PricePerItem float64   `db:"price_per_item" json:"price_per_item"`
	TotalPrice   float64   `db:"total_price" json:"total_price"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`

	Variant *ProductVariant `db:"-" json:"variant,omitempty"`
	Product *Product        `db:"-" json:"product,omitempty"`
}

type OrderStatus struct {
	ID        int                 `db:"id" json:"id"`
	OrderID   int                 `db:"order_id" json:"order_id"`
	Status    string              `db:"status" json:"status"`
	Notes     nullable.NullString `db:"notes" json:"notes"`
	CreatedBy string              `db:"created_by" json:"created_by"`
	CreatedAt time.Time           `db:"created_at" json:"created_at"`
}

type CheckoutRequest struct {
	AddressID   int    `json:"address_id"`
	Notes       string `json:"notes"`
	CartItemIDs []int  `json:"cart_items_ids"`
}

type CheckoutResponse struct {
	Order Order `json:"order"`
}

type UpdateOrderStatusReq struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}
