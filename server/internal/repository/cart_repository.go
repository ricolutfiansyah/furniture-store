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
	err := r.db.GetContext(ctx, &cart, query)

	if err == sql.ErrNoRows {
		insertQuery := `INSERT INTO carts (user_id) VALUES (?)`
		result, err := r.db.ExecContext(ctx, insertQuery, userID)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		cart.ID = int(id)
		cart.UserID = userID
		return &cart, nil
	}

	return &cart, nil
}
