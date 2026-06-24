package repository

import (
	"context"
	"database/sql"
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
	var cart domain.Cart
	query := `SELECT * FROM carts WHERE user_id = ?`
	err := r.db.GetContext(ctx, &cart, query, userID)

	if err == sql.ErrNoRows {
		insertQuery := `INSERT INTO carts (user_id) VALUES (?)`
		result, err := r.db.ExecContext(ctx, insertQuery, userID)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		cart.ID = int(id)
		cart.UserID = userID
		return &cart, nil
	}

	if err != nil {
		return nil, err
	}

	return &cart, nil
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, userID int) (*domain.Cart, error) {
	cart, err := r.GetCartWithItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	var items []domain.CartItem
	// query := `SELECT ci.*,
	// pv.*,
	// p.id, as product_id, p.name as product_name, p.slug as product_slug, p.base_price
	// FROM cart_items ci
	// JOIN prodict_variants pv ON ci.variant_id = pv.id
	// JOIN products p ON pv.product_id = p.id
	// WHERE ci.cart_id = ?
	// `

	itemsQuery := `SELECT * FROM cart_items WHERE id = ?`
	err = r.db.SelectContext(ctx, &items, itemsQuery, userID)
	if err != nil {
		return nil, err
	}

	cart.Items = items
	return cart, nil
}

func (r *cartRepository) UpdateItemRepository(ctx context.Context, cartItemID int, quantity int) error {
	query := `UPDATE cart_items SET quantity = ? WHERE id = ? AND quantity > 0`
	_, err := r.db.ExecContext(ctx, query, cartItemID)
	return err
}
