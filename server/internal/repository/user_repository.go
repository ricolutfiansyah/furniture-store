package repository

import (
	"context"
	"database/sql"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

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
	query := `SELECT * FROM users WHERE email = ?`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindById(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE id = ?`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, nil
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET full_name = ?, phone = ?, address = ?, updatedAt = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, user.FullName, user.Phone, user.Address, user.ID)
	return err
}
