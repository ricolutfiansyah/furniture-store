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

func (r *orderRepository) CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, variant_id, quantity, price_per_item, total_price)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := tx.ExecContext(ctx, query, item.OrderID, item.VariantID, item.Quantity, item.PricePerItem, item.TotalPrice)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = int(id)

	selectQuery := `SELECT created_at FROM order_items WHERE id = ?`
	err = tx.QueryRowContext(ctx, selectQuery, id).Scan(&item.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *orderRepository) CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status, notes, createdBy string) error {
	query := `
        INSERT INTO order_statuses (order_id, status, notes, created_by)
        VALUES (?, ?, ?, ?)
    `
	_, err := tx.ExecContext(ctx, query, orderID, status, notes, createdBy)
	return err
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	var orders []domain.Order
	query := `SELECT * FROM orders WHERE user_id = ?`
	err := r.db.SelectContext(ctx, &orders, query, userID)
	return orders, err
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID int) (*domain.Order, error) {
	var order domain.Order
	query := `SELECT * FROM orders WHERE id = ?`
	err := r.db.GetContext(ctx, &order, query, orderID)
	return &order, err
}

func (r *orderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	query := `SELECT * FROM order_items WHERE order_id = ?`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	return items, err
}

func (r *orderRepository) GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
	var statuses []domain.OrderStatus
	query := `SELECT * FROM order_statuses WHERE id = ? ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &statuses, query, orderID)
	return statuses, err
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	query := `UPDATE orders SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	return err
}
