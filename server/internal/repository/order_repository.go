package repository

import (
	"context"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrderWithTx(ctx context.Context, tx *sqlx.Tx, order *domain.Order) error {
	query := `
		INSERT INTO orders
			(user_id, order_number, total_amount, shipping_cost, tax, grand_total, 
			status, shipping_address, payment_method, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.ExecContext(ctx, query,
		order.UserID,
		order.OrderNumber,
		order.TotalAmount,
		order.ShippingCost,
		order.Tax,
		order.GrandTotal,
		order.Status,
		order.ShippingAddress,
		order.PaymentMethod,
		order.Notes,
	)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	order.ID = int(id)
	return nil
}
