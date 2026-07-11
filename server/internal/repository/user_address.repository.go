package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type addressRepository struct {
	db *sqlx.DB
}

func NewAddressRepository(db *sqlx.DB) *addressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *domain.UserAddress) error {
	const query = `
				INSERT INTO user_addresses 
					(user_id, label, recipient_name, phone, province, city, district, postal_code, address_line, is_default)
				VALUES
					(:user_id, :label, :recipient_name, :phone, :province, :city, :district, :postal_code, :address_line, :is_default)
			`
	result, err := r.db.NamedExecContext(ctx, query, address)
	if err != nil {
		return fmt.Errorf("insert address: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id %w", err)
	}
	address.ID = int(id)

	return nil
}

func (r *addressRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	const query = `SELECT COUNT(*) FROM user_addresses WHERE user_id = ?`

	var count int
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("count addresses: %w", err)
	}
	return count, nil
}

func (r *addressRepository) GetByID(ctx context.Context, id, userID int) (*domain.UserAddress, error) {
	const query = `
				SELECT id, user_id, label, recipient_name, phone, province, city, district,
						postal_code, address_line, is_default, created_at, updated_at
				FROM user_addresses
				WHERE id = ? AND user_id = ?
				FOR UPDATE
			`

	var address domain.UserAddress
	err := r.db.GetContext(ctx, &address, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("get address by id: %w", err)
	}

	return &address, nil
}

func (r *addressRepository) GetByIDTx(ctx context.Context, tx *sqlx.Tx, id, userID int) (*domain.UserAddress, error) {
	const query = `
				SELECT id, user_id, label, recipient_name, phone, province, city, district,
						postal_code, address_line, is_default, created_at, updated_at
				FROM user_addresses
				WHERE id = ? AND user_id = ?
				FOR UPDATE
			`

	var address domain.UserAddress
	err := tx.GetContext(ctx, &address, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("get address by id (tx): %w", err)
	}

	return &address, nil
}

func (r *addressRepository) ListByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.UserAddress, error) {
	const query = `
				SELECT id, user_id, label, recipient_name, phone, province, city, district,
						postal_code, address_line, is_default, created_at, updated_at
				FROM user_addresses		
				WHERE user_id = ?
			`

	addresses := []domain.UserAddress{}
	if err := tx.SelectContext(ctx, &addresses, query, userID); err != nil {
		return nil, fmt.Errorf("list addresses (tx): %w", err)
	}

	return addresses, nil
}

func (r *addressRepository) Update(ctx context.Context, address *domain.UserAddress) error {
	const query = `
				UPDATE user_addresses
				SET label = :label, recipient_name = :recipient_name, phone = :phone,
					province = :province, city = :city, district := district,
					postal_code = :postal_code, address_line = :address_line
				WHERE id = :id AND user_id = :user_id
			`

	result, err := r.db.NamedExecContext(ctx, query, address)
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAddressNotFound
	}

	return nil
}

func (r *addressRepository) DeleteTx(ctx context.Context, tx *sqlx.Tx, id, userID int) error {
	const query = `DELETE FROM user_addresses WHERE id = ? AND user_id = ?`

	result, err := tx.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected (tx): %w", err)
	}
	if rowsAffected == 0 {
		return ErrAddressNotFound
	}

	return nil
}

func (r *addressRepository) UnsetDefaultByUserID(ctx context.Context, tx *sqlx.Tx, userID int) error {
	const query = `UPDATE user_addresses SET is_default = FALSE WHERE user_id = ?`

	if _, err := tx.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("unset default addresses: %w", err)
	}

	return nil
}

func (r *addressRepository) SetDefault(ctx context.Context, tx *sqlx.Tx, id, userID int) error {
	query := `UPDATE user_addresses SET is_default = TRUE WHERE id = ? AND user_id = ?`
	result, err := tx.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("set default address: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrAddressNotFound
	}
	return nil
}
