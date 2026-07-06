package domain

import (
	"furniture-api/internal/nullable"
	"time"
)

type UserAddress struct {
	ID            string              `db:"id" json:"id"`
	UserID        int                 `db:"user_id" json:"user_id"`
	Label         nullable.NullString `db:"label" json:"label"`
	RecipientName string              `db:"recipient_name" json:"recipient_name"`
	Phone         string              `db:"phone" json:"phone"`
	Province      string              `db:"province" json:"province"`
	City          string              `db:"city" json:"city"`
	District      string              `db:"district" json:"district"`
	PostalCode    string              `db:"postal_code" json:"postal_code"`
	AddressLine   string              `db:"address_line" json:"address_line"`
	IsDefault     bool                `db:"is_default" json:"is_default"`
	CreatedAt     time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `db:"updated_at" json:"updated_at"`
}
