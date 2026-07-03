package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

var ErrUserNotFound = errors.New("user not found")

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (public_id, email, password_hash, full_name, phone, address, role) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		user.PublicID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.Address,
		user.Role,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)

	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	const query = `SELECT id, public_id, email, password_hash, full_name, phone, address, role, is_active, created_at, updated_at FROM users WHERE email = ?`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) FindById(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	const query = `SELECT id, public_id, email, password_hash, full_name, phone, address, role, is_active, created_at, updated_at FROM users WHERE id = ?`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `UPDATE users SET full_name = ?, phone = ?, address = ?, updated_at = NOW() WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		user.FullName,
		user.Phone,
		user.Address,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return err
}
