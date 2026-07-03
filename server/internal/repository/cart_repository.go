package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type cartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) *cartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreateCart(ctx context.Context, userID int) (*domain.Cart, error) {
	cart, err := r.findByUserID(ctx, userID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, ErrCartNotFound) {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	insertQuery := `INSERT INTO carts (user_id) VALUES (?)`
	if _, err := r.db.ExecContext(ctx, insertQuery, userID); err != nil {
		if isDuplicateKeyError(err, "user_id") {
			return r.findByUserID(ctx, userID)
		}
		return nil, fmt.Errorf("create cart: %w", err)
	}

	return r.findByUserID(ctx, userID)
}

func (r *cartRepository) findByUserID(ctx context.Context, userID int) (*domain.Cart, error) {
	const query = `SELECT id, user_id, created_at, updated_at FROM carts WEHERE user_id = ?`

	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("find cart by user id: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, userID int) (*domain.Cart, error) {
	cart, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	var items []domain.CartItem

	itemsQuery := `SELECT * FROM cart_items WHERE cart_id = ?`
	err = r.db.SelectContext(ctx, &items, itemsQuery, cart.ID)
	if err != nil {
		return nil, err
	}

	cart.Items = items
	return cart, nil
}

func (r *cartRepository) GetCartItemsByUserID(ctx context.Context, userID int) ([]domain.CartItem, error) {
	var items []domain.CartItem
	query := `SELECT ci.* FROM cart_items ci JOIN carts c ON ci.cart_id = c.id WHERE c.user_id = ?`
	err := r.db.SelectContext(ctx, &items, query, userID)
	return items, err
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID int) (*domain.Cart, error) {
	var cart domain.Cart
	query := `SELECT * FROM carts WHERE user_id = ?`
	err := r.db.GetContext(ctx, &cart, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &cart, err
}

func (r *cartRepository) AddItem(ctx context.Context, cartID int, variantID int, quantity int, PriceAtTime float64) error {
	query := `
		INSERT INTO cart_items (cart_id, variant_id, quantity, price_at_time) 
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE quantity = quantity + ?
	`
	_, err := r.db.ExecContext(ctx, query, cartID, variantID, quantity, PriceAtTime, quantity)
	return err
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartItemID, quantity int) error {
	query := `UPDATE cart_items SET quantity = ? WHERE id = ? AND quantity > 0`
	_, err := r.db.ExecContext(ctx, query, quantity, cartItemID)
	return err
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartItemID int) error {
	query := `DELETE FROM cart_items WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, cartItemID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID int) error {
	query := `DELETE FROM cart_items WHERE cart_id = ?`
	_, err := r.db.ExecContext(ctx, query, cartID)
	return err
}
