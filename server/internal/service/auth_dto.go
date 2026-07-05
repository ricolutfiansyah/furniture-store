package service

import (
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
)

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

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:       user.PublicID,
		Email:    user.Email,
		FullName: user.FullName,
		Phone:    user.Phone,
		Address:  user.Address,
		Role:     user.Role,
	}
}
