package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository/queries"
	"time"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := r.db.NamedExecContext(ctx, queries.UserInsert, user)
	if err != nil {
		if isDuplicateKeyError(err, "email") {
			return domain.ErrEmailAlreadyRegistered
		}
		if isDuplicateKeyError(err, "public_id") {
			return fmt.Errorf("public id collision: %w", err)
		}
		return fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	user.ID = int(id)

	var timestamps struct {
		IsActive  bool      `db:"is_active"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
	if err := r.db.GetContext(ctx, &timestamps, queries.UserSelectTimestamps, user.ID); err != nil {
		return fmt.Errorf("fetch created user timestamps: %w", err)
	}
	user.IsActive = timestamps.IsActive
	user.CreatedAt = timestamps.CreatedAt
	user.UpdatedAt = timestamps.UpdatedAt

	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, queries.UserSelectByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) FindById(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, queries.UserSelectByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	return &user, nil
}

func (r *userRepository) FindByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, queries.UserSelectByPublicID, publicID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by public ID: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.NamedExecContext(ctx, queries.UserUpdate, user)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return err
}
