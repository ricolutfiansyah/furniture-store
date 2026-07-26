package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository/queries"

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
	if !errors.Is(err, domain.ErrCartNotFound) {
		return nil, fmt.Errorf("get cart: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, queries.CartInsert, userID); err != nil {
		if isDuplicateKeyError(err, "user_id") {
			return r.findByUserID(ctx, userID)
		}
		return nil, fmt.Errorf("create cart: %w", err)
	}

	return r.findByUserID(ctx, userID)
}

func (r *cartRepository) findByUserID(ctx context.Context, userID int) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, queries.CartSelectByUserID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCartNotFound
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

	items := []domain.CartItem{}
	if err = r.db.SelectContext(ctx, &items, queries.CartItemSelectByCartID, cart.ID); err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}

	cart.Items = items
	return cart, nil
}

func (r *cartRepository) GetCartItemsByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.CartItem, error) {
	items := []domain.CartItem{}
	if err := tx.SelectContext(ctx, &items, queries.CartItemSelectByUserIDTx, userID); err != nil {
		return nil, fmt.Errorf("get cart items by user id: %w", err)
	}

	return items, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID, variantID, quantity int, PriceAtTime float64) error {
	_, err := r.db.ExecContext(ctx, queries.CartItemInsert, cartID, variantID, quantity, PriceAtTime)
	if err != nil {
		return fmt.Errorf("add item to cart: %w", err)
	}

	return nil
}

func (r *cartRepository) GetCartItem(ctx context.Context, cartID, variantID int) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.GetContext(ctx, &item, queries.CartItemSelectByCartAndVariant, cartID, variantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("find cart item %w", err)
	}
	return &item, nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, userID, cartItemID, quantity int) error {
	result, err := r.db.ExecContext(ctx, queries.CartItemUpdateQuantity, quantity, cartItemID, userID)
	if err != nil {
		return fmt.Errorf("update item quantity: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update item quantity rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID, cartItemID int) error {
	result, err := r.db.ExecContext(ctx, queries.CartItemDelete, cartItemID, userID)
	if err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove cart item rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) GetCartItemsByIDsTx(ctx context.Context, tx *sqlx.Tx, userID int, itemIDs []int) ([]domain.CartItem, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(queries.CartItemSelectByIDsTx, userID, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	query = tx.Rebind(query)
	items := []domain.CartItem{}
	if err := tx.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("get cart items by ids: %w", err)
	}

	return items, nil
}

func (r *cartRepository) RemoveCartItemsWithTx(ctx context.Context, tx *sqlx.Tx, cartID int, itemIDs []int) error {
	if len(itemIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(queries.CartItemDeleteTx, cartID, itemIDs)
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	query = tx.Rebind(query)

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("remove specific cart items: %w", err)
	}

	return nil
}

func (r *cartRepository) RemoveItems(ctx context.Context, userID int, itemIDs []int) error {
	if len(itemIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(queries.CartItemDeleteBulk, userID, itemIDs)
	if err != nil {
		return fmt.Errorf("build bulk delete query: %w", err)
	}

	query = r.db.Rebind(query)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("bulk remove cart items: %w", err)
	}

	return nil
}

func (r *cartRepository) GetCartItemByID(ctx context.Context, userID, cartItemID int) (*domain.CartItem, error) {
	var item domain.CartItem
	if err := r.db.GetContext(ctx, &item, queries.CartItemSelectByID, cartItemID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("get cart item by id: %w", err)
	}
	return &item, nil
}
