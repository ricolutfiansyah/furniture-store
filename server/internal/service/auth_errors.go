package service

import "errors"

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
)
