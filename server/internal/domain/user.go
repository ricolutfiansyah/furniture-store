package domain

import (
	"furniture-api/internal/nullable"
	"time"
)

type User struct {
	ID           int                 `db:"id" json:"-"`
	PublicID     string              `db:"public_id" json:"id"`
	Email        string              `db:"email" json:"email"`
	PasswordHash string              `db:"password_hash" json:"-"`
	FullName     nullable.NullString `db:"full_name" json:"full_name"`
	Phone        nullable.NullString `db:"phone" json:"phone"`
	Address      nullable.NullString `db:"address" json:"address"`
	Role         string              `db:"role" json:"role"`
	IsActive     bool                `db:"is_active" json:"is_active"`
	CreatedAt    time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time           `db:"updated_at" json:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       string              `json:"id"`
	Email    string              `json:"email"`
	FullName nullable.NullString `json:"full_name"`
	Phone    nullable.NullString `json:"phone"`
	Address  nullable.NullString `json:"address"`
	Role     string              `json:"role"`
}

func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:       user.PublicID,
		Email:    user.Email,
		FullName: user.FullName,
		Phone:    user.Phone,
		Address:  user.Address,
		Role:     user.Role,
	}
}
