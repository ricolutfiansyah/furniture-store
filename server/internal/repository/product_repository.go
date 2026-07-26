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

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetActive(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	products := []domain.Product{}
	err := r.db.SelectContext(ctx, &products, queries.ProductSelectActive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get active products: %w", err)
	}

	return products, nil
}

func (r *productRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, queries.ProductCountActive)
	if err != nil {
		return 0, fmt.Errorf("count active products: %w", err)
	}

	return count, nil
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, queries.ProductSelectBySlug, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by slug: %w", err)
	}

	return &product, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, queries.ProductSelectByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by id: %w", err)
	}

	return &product, nil
}

func (r *productRepository) GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error) {
	variants := []domain.ProductVariant{}
	err := r.db.SelectContext(ctx, &variants, queries.VariantSelectByProductID, productID)
	if err != nil {
		return nil, fmt.Errorf("get variants by product id: %w", err)
	}

	return variants, nil
}

func (r *productRepository) GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	images := []domain.ProductImage{}
	err := r.db.SelectContext(ctx, &images, queries.ImageSelectByProductID, productID)
	if err != nil {
		return nil, fmt.Errorf("get images by product id: %w", err)
	}

	return images, nil
}

func (r *productRepository) GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error) {
	var category domain.Category
	err := r.db.GetContext(ctx, &category, queries.CategorySelectByID, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}

	return &category, nil
}

func (r *productRepository) GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error) {
	var variant domain.ProductVariant
	err := r.db.GetContext(ctx, &variant, queries.VariantSelectByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrVariantNotFound
		}
		return nil, fmt.Errorf("get variant by id: %w", err)
	}
	return &variant, nil
}

func (r *productRepository) DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error {
	result, err := tx.ExecContext(ctx, queries.VariantDecreaseStock, quantity, variantID, quantity)
	if err != nil {
		return fmt.Errorf("decrease stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("decrease stock rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrInsufficientStock
	}

	return nil
}
