package domain

import "time"

type Cart struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Items []CartItem `db:"-" json:"items,omitempty"`
}

type CartItem struct {
	ID          int       `db:"id" json:"id"`
	CartID      int       `db:"cart_id" json:"cart_id"`
	VariantID   int       `db:"variant_id" json:"variant_id"`
	Quantity    int       `db:"quantity" json:"quantity"`
	PriceAtTime float64   `db:"price_at_time" json:"price_at_time"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	Variant *ProductVariant `db:"-" json:"variant,omitempty"`
	Product *Product        `db:"-" json:"product,omitempty"`
}

type AddToCartRequest struct {
	VariantID int `json:"variant_id"`
	Quantity  int `json:"quantity"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type BulkRemoveCartItems struct {
	CartItemIDs []int `json:"cart_item_ids"`
}
