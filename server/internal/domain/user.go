package domain

import "time"

type User struct {
	ID           int       `db:"id" json:"-"`
	PublicID     string    `db:"public_id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	FullName     string    `db:"full_name" json:"full_name"`
	Phone        string    `db:"phone" json:"phone"`
	Address      string    `db:"address" json:"address"`
	Role         string    `db:"role" json:"role"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
