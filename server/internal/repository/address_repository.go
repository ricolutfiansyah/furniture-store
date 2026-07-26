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

type addressRepository struct {
	db *sqlx.DB
}

func NewAddressRepository(db *sqlx.DB) *addressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *domain.UserAddress) error {
	result, err := r.db.NamedExecContext(ctx, queries.AddressInsert, address)
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
	var count int
	if err := r.db.GetContext(ctx, &count, queries.AddressCountByUserID, userID); err != nil {
		return 0, fmt.Errorf("count addresses: %w", err)
	}
	return count, nil
}

func (r *addressRepository) GetByID(ctx context.Context, id, userID int) (*domain.UserAddress, error) {
	var address domain.UserAddress
	err := r.db.GetContext(ctx, &address, queries.AddressSelectByID, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("get address by id: %w", err)
	}

	return &address, nil
}

func (r *addressRepository) GetByIDTx(ctx context.Context, tx *sqlx.Tx, id, userID int) (*domain.UserAddress, error) {
	var address domain.UserAddress
	err := tx.GetContext(ctx, &address, queries.AddressSelectByIDForUpdate, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("get address by id (tx): %w", err)
	}

	return &address, nil
}

func (r *addressRepository) ListByUserID(ctx context.Context, userID int) ([]domain.UserAddress, error) {
	addresses := []domain.UserAddress{}
	if err := r.db.SelectContext(ctx, &addresses, queries.AddressSelectByUserID, userID); err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}

	return addresses, nil
}

func (r *addressRepository) ListByUserIDTx(ctx context.Context, tx *sqlx.Tx, userID int) ([]domain.UserAddress, error) {
	addresses := []domain.UserAddress{}
	if err := tx.SelectContext(ctx, &addresses, queries.AddressSelectByUserIDTx, userID); err != nil {
		return nil, fmt.Errorf("list addresses (tx): %w", err)
	}

	return addresses, nil
}

func (r *addressRepository) Update(ctx context.Context, address *domain.UserAddress) error {
	result, err := r.db.NamedExecContext(ctx, queries.AddressUpdate, address)
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

func (r *addressRepository) DeleteTx(ctx context.Context, tx *sqlx.Tx, id, userID int) error {
	result, err := tx.ExecContext(ctx, queries.AddressDelete, id, userID)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected (tx): %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

func (r *addressRepository) UnsetDefaultByUserID(ctx context.Context, tx *sqlx.Tx, userID int) error {
	if _, err := tx.ExecContext(ctx, queries.AddressUnsetDefault, userID); err != nil {
		return fmt.Errorf("unset default addresses: %w", err)
	}

	return nil
}

func (r *addressRepository) SetDefault(ctx context.Context, tx *sqlx.Tx, id, userID int) error {
	result, err := tx.ExecContext(ctx, queries.AddressSetDefault, id, userID)
	if err != nil {
		return fmt.Errorf("set default address: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}
