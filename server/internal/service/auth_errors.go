package service

import "errors"

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)
