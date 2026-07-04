package repository

import "errors"

var (
	ErrDuplicateOrderNumber = errors.New("order number already exist")
)
