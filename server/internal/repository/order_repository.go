package repository

import "github.com/jmoiron/sqlx"

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *orderRepository {
	return &orderRepository{db: db}
}
