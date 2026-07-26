package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/nullable"
	"furniture-api/internal/repository/queries"
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
	result, err := tx.NamedExecContext(ctx, queries.OrderInsert, order)
	if err != nil {
		if isDuplicateKeyError(err, "order_number") {
			return domain.ErrDuplicateOrderNumber
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
	if err := tx.GetContext(ctx, &created, queries.OrderSelectTimestamps, id); err != nil {
		return fmt.Errorf("fetch created order timestamps: %w", err)
	}

	order.ID = int(id)
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt

	return nil
}

func (r *orderRepository) CreateOrderItemWithTx(ctx context.Context, tx *sqlx.Tx, item *domain.OrderItem) error {
	result, err := tx.NamedExecContext(ctx, queries.OrderItemInsert, item)
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("create order item last insert id: %w", err)
	}
	item.ID = int(id)

	if err = tx.GetContext(ctx, &item.CreatedAt, queries.OrderItemSelectTimestamp, id); err != nil {
		return fmt.Errorf("fetch created order item timestamp: %w", err)
	}

	return nil
}

func (r *orderRepository) CreateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, orderID int, status string, notes nullable.NullString, createdBy string) error {
	if _, err := tx.ExecContext(ctx, queries.OrderStatusInsert, orderID, status, notes, createdBy); err != nil {
		return fmt.Errorf("create order status: %w", err)
	}

	return nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	orders := []domain.Order{}
	if err := r.db.SelectContext(ctx, &orders, queries.OrderSelectByUserID, userID); err != nil {
		return nil, fmt.Errorf("get order by user id: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	var order domain.Order
	if err := r.db.GetContext(ctx, &order, queries.OrderSelectByID, orderID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) GetOrderByIDForAdmin(ctx context.Context, orderID int) (*domain.Order, error) {
	var order domain.Order
	if err := r.db.GetContext(ctx, &order, queries.OrderSelectByIDForAdmin, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id for admin: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]domain.OrderItem, error) {
	items := []domain.OrderItem{}
	if err := r.db.SelectContext(ctx, &items, queries.OrderItemSelectByOrderID, orderID); err != nil {
		return nil, fmt.Errorf("get order items by order id: %w", err)
	}

	return items, nil
}

func (r *orderRepository) GetOrderStatusesByOrderID(ctx context.Context, orderID int) ([]domain.OrderStatus, error) {
	statuses := []domain.OrderStatus{}
	if err := r.db.SelectContext(ctx, &statuses, queries.OrderStatusSelectByOrderID, orderID); err != nil {
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
	query := queries.OrderStatusUpdate

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
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *orderRepository) GetOrderStatusForUpdate(ctx context.Context, tx *sqlx.Tx, orderID int) (string, error) {
	var status string
	if err := tx.GetContext(ctx, &status, queries.OrderStatusSelectForUpdate, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrOrderNotFound
		}
		return "", fmt.Errorf("get order status for update: %w", err)
	}

	return status, nil
}

func (r *orderRepository) GetOrderSummaries(ctx context.Context, orderIDs []int) (map[int]domain.OrderSummary, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(queries.OrderSummarySelect, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("build query summaries: %w", err)
	}

	query = r.db.Rebind(query)

	var summaries []domain.OrderSummary
	if err := r.db.SelectContext(ctx, &summaries, query, args...); err != nil {
		return nil, fmt.Errorf("get orders summaries: %w", err)
	}

	result := make(map[int]domain.OrderSummary)
	for _, s := range summaries {
		result[s.OrderID] = s
	}

	return result, nil
}
