package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"time"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrderWithTx(ctx context.Context, tx *sqlx.Tx, order *domain.Order) error {
	const query = `
		INSERT INTO orders
			(user_id, order_number, total_amount, shipping_cost, tax, grand_total, 
			status, shipping_address, payment_method, notes
		) VALUES (:user_id, :order_number, :total_amount, :shipping_cost, :tax, :grand_total, 
		 	:status, :shipping_address, :payment_method, :notes)
	`

	result, err := tx.NamedExecContext(ctx, query, order)
	if err != nil {
		if isDuplicateKeyError(err, "order_number") {
			return ErrDuplicateOrderNumber
		}
		return fmt.Errorf("create order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("create order last insert: %w", err)
	}

	var created struct {
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
	const selectQuery = `SELECT created_at, updated_at FROM orders WHERE id = ?`
	if err := tx.GetContext(ctx, &created, selectQuery, id); err != nil {
		return fmt.Errorf("fetch created order timestamps: %w", err)
	}

	order.ID = int(id)
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt

	return nil
}

func (r *orderRepository) CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error {
	const query = `
		INSERT INTO order_items (order_id, variant_id, quantity, price_per_item, total_price)
		VALUES (:order_id, :variant_id, :quantity, :price_per_item, :total_price)
	`
	result, err := tx.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("create order item last insert id: %w", err)
	}
	item.ID = int(id)

	const selectQuery = `SELECT created_at FROM order_items WHERE id = ?`
	if err = tx.GetContext(ctx, &item.CreatedAt, selectQuery, id); err != nil {
		return fmt.Errorf("fetch created order item timestamp: %w", err)
	}

	return nil
}

func (r *orderRepository) CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, notes nullable.NullString, createdBy string) error {
	const query = `
        	INSERT INTO order_statuses (order_id, status, notes, created_by)
        	VALUES (?, ?, ?, ?)
    `
	if _, err := tx.ExecContext(ctx, query, orderID, status, notes, createdBy); err != nil {
		return fmt.Errorf("create order status: %w", err)
	}

	return nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	const query = `
				SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
					status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
					notes, created_at, updated_at 
				FROM orders WHERE user_id = ?
				ORDER BY created_at DESC`

	orders := []domain.Order{}
	if err := r.db.SelectContext(ctx, &orders, query, userID); err != nil {
		return nil, fmt.Errorf("get order by user id: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	const query = `
				SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
					status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
					notes, created_at, updated_at 
				FROM orders WHERE id = ? AND user_id = ?`

	var order domain.Order
	if err := r.db.GetContext(ctx, &order, query, orderID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	return &order, nil
}

// ! ADMIN ONLY PURPOSE
func (r *orderRepository) GetOrderByIDForAdmin(ctx context.Context, orderID int) (*domain.Order, error) {
	const query = `
		SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
			status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
			notes, created_at, updated_at
		FROM orders WHERE id = ?`

	var order domain.Order
	if err := r.db.GetContext(ctx, &order, query, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id for admin: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
	const query = `
				SELECT id, order_id, variant_id, quantity, price_per_item, total_price, created_at 
				FROM order_items WHERE order_id = ?`

	items := []domain.OrderItem{}
	if err := r.db.SelectContext(ctx, &items, query, orderID); err != nil {
		return nil, fmt.Errorf("get order items by order id: %w", err)
	}

	return items, nil
}

func (r *orderRepository) GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
	const query = `
				SELECT id, order_id, status, notes, created_by, created_at 
				FROM order_statuses WHERE order_id = ? ORDER BY created_at ASC`

	statuses := []domain.OrderStatus{}
	if err := r.db.SelectContext(ctx, &statuses, query, orderID); err != nil {
		return nil, fmt.Errorf("get order statuses by order id: %w", err)
	}

	return statuses, nil
}

var allowedTimestampColumns = map[string]bool{
	"paid_at":      true,
	"shipped_at":   true,
	"delivered_at": true,
}

func (r *orderRepository) UpdateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status, timestampColumn string) error {
	query := `UPDATE orders SET status = ?`

	if timestampColumn != "" {
		if !allowedTimestampColumns[timestampColumn] {
			return fmt.Errorf("update order status: invalid timestamp column %q", timestampColumn)
		}
		query += `, ` + timestampColumn + ` = NOW()`
	}
	query += ` WHERE id = ?`

	result, err := tx.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update order status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func (r *orderRepository) GetOrderStatusForUpdate(ctx context.Context, tx *sqlx.Tx, orderID int) (string, error) {
	const query = `SELECT status FROM orders WHERE id = ? FOR UPDATE`

	var status string
	if err := tx.GetContext(ctx, &status, query, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrOrderNotFound
		}
		return "", fmt.Errorf("get order status for update: %w", err)
	}

	return status, nil
}

type OrderSummaryDTO struct {
	OrderID     int    `db:"order_id"`
	TotalItems  int    `db:"total_items"`
	VariantName string `db:"variant_name"`
	ImageURL    string `db:"image_url"`
}

func (r *orderRepository) GetOrderSummaries(ctx context.Context, orderIDs []int) (map[int]OrderSummaryDTO, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		SELECT oi.order_id, (SELECT COUNT(*) FROM order_items WHERE order_id = oi.order_id) as total_items, pv.variant_name,
			COALESCE(pi.image_url, '') as image_url
		FROM order_items oi
		JOIN product_variants pv ON oi.variant_id = pv.id
		LEFT JOIN product_images pi ON pv.product_id = pi.product_id AND pi.is_primary = TRUE
		WHERE oi.id IN (SELECT MIN(id) FROM order_items WHERE order_id IN (?) GROUP BY order_id
		)
	`, orderIDs)

	if err != nil {
		return nil, fmt.Errorf("build query summaries: %w", err)
	}

	query = r.db.Rebind(query)

	var summaries []OrderSummaryDTO
	if err := r.db.SelectContext(ctx, &summaries, query, args...); err != nil {
		return nil, fmt.Errorf("get orders summaries: %w", err)
	}

	result := make(map[int]OrderSummaryDTO)
	for _, s := range summaries {
		result[s.OrderID] = s
	}

	return result, nil
}
