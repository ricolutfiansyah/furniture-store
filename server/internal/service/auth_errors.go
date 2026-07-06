package service

import "errors"

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrInvalidInput           = errors.New("invalid input")

	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")

	ErrInvalidEmail = errors.New("invalid email format")
)
