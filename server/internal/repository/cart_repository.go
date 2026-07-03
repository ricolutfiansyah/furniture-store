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

	const insertQuery = `INSERT INTO carts (user_id) VALUES (?)`
	if _, err := r.db.ExecContext(ctx, insertQuery, userID); err != nil {
		if isDuplicateKeyError(err, "user_id") {
			return r.findByUserID(ctx, userID)
		}
		return nil, fmt.Errorf("create cart: %w", err)
	}

	return r.findByUserID(ctx, userID)
}

func (r *cartRepository) findByUserID(ctx context.Context, userID int) (*domain.Cart, error) {
	const query = `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = ?`

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
		return nil, fmt.Errorf("get or create cart: %w", err)
	}

	var items []domain.CartItem
	const itemsQuery = `
					SELECT id, cart_id, variant_id, quantity, price_at_time, created_at, updated_at 
					FROM cart_items 
					WHERE cart_id = ?`

	if err = r.db.SelectContext(ctx, &items, itemsQuery, cart.ID); err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	cart.Items = items
	return cart, nil
}

func (r *cartRepository) GetCartItemsByUserIDTx(ctx context.Context, tx sqlx.Tx, userID int) ([]domain.CartItem, error) {
	const query = `
				SELECT ci.id, ci.cart_id, ci/variant_id, ci.quantity, ci.price_at_time, ci.created_at 
				FROM cart_items ci 
				JOIN carts c ON ci.cart_id = c.id 
				WHERE c.user_id = ?`

	var items []domain.CartItem
	if err := tx.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, fmt.Errorf("get cart items by user id: %w", err)
	}

	return items, nil
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

func (r *cartRepository) AddItem(ctx context.Context, cartID int, variantID int, quantity int, PriceAtTime float64) (*domain.CartItem, error) {
	const query = `
		INSERT INTO cart_items (cart_id, variant_id, quantity, price_at_time) 
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			quantity = quantity + VALUES(quantity)
	`
	_, err := r.db.ExecContext(ctx, query, cartID, variantID, quantity, PriceAtTime)
	if err != nil {
		return nil, fmt.Errorf("add item to cart: %w", err)
	}

	return r.findByCartAndVariant(ctx, cartID, variantID)
}

func (r *cartRepository) findByCartAndVariant(ctx context.Context, cartID, variantID int) (*domain.CartItem, error) {
	const query = `
					SELECT id, cart_id, variant_id, quantity, price_at_time, created_at, updated_at
					FROM cart_items
					WHERE cart_id = ? AND variant_id = ?`

	var item domain.CartItem
	err := r.db.GetContext(ctx, &item, query, cartID, variantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("find cart item %w", err)
	}
	return &item, nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartItemID, quantity int) error {
	const query = `UPDATE cart_items SET quantity = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, quantity, cartItemID)
	if err != nil {
		return fmt.Errorf("update item quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update item quantity rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartItemID int) error {
	const query = `DELETE FROM cart_items WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, cartItemID)
	if err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove cart item rows affected")
	}
	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID int) error {
	query := `DELETE FROM cart_items WHERE cart_id = ?`
	_, err := r.db.ExecContext(ctx, query, cartID)
	return err
}
