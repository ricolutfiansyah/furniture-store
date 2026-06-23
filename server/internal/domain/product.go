package domain

import "time"

type Category struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Slug        string     `db:"slug" json:"slug"`
	Description string     `db:"description" json:"description"`
	ParentID    NullString `db:"parent_id" json:"parent_id"`
	ImageURL    string     `db:"image_url" json:"image_url"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type Product struct {
	ID          int       `db:"id" json:"id"`
	CategoryID  int       `db:"category_id" json:"category_id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	Description string    `db:"description" json:"description"`
	BasePrice   float64   `db:"base_price" json:"base_price"`
	SKU         string    `db:"sku" json:"sku"`
	WeightKg    float64   `db:"weight_kg" json:"weight_kg"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	Views       int       `db:"views" json:"views"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	Category *Category        `db:"-" json:"category,omitempty"`
	Variants []ProductVariant `db:"-" json:"variants,omitempty"`
	Images   []ProductImage   `db:"-" json:"images,omitempty"`
}

type ProductVariant struct {
	ID              int       `db:"id" json:"id"`
	ProductID       int       `db:"product_id" json:"product_id"`
	VariantName     string    `db:"variant_name" json:"variant_name"`
	Attributes      JSON      `db:"attributes" json:"attributes"`
	AdditionalPrice float64   `db:"additional_price" json:"additional_price"`
	StockQuantity   int       `db:"stock_quantity" json:"stock_quantity"`
	SKUVariant      string    `db:"sku_variant" json:"sku_variant"`
	WeightKg        float64   `db:"weight_kg" json:"weight_kg"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type ProductImage struct {
	ID        int       `db:"id" json:"id"`
	ProductID int       `db:"product_id" json:"product_id"`
	VariantID int       `db:"variant_id" json:"variant_id,omitempty"`
	ImageURL  string    `db:"image_url" json:"image_url"`
	IsPrimary bool      `db:"is_primary" json:"is_primary"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
