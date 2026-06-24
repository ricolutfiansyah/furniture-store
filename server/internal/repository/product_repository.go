package repository

import (
	"context"
	"database/sql"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	var products []domain.Product
	query := `SELECT * FROM products WHERE is_active = TRUE ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &products, query, limit, offset)
	return products, err
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var product domain.Product
	query := `SELECT * FROM products WHERE slug = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &product, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error) {
	var variants []domain.ProductVariant
	query := `SELECT * FROM product_variants WHERE product_id = ? AND is_active = TRUE`
	err := r.db.SelectContext(ctx, &variants, query, productID)
	return variants, err
}

func (r *productRepository) GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	var images []domain.ProductImage
	query := `SELECT * FROM product_images WHERE product_id = ? ORDER_ID = ? is_active DESC, sort_order ASC`
	err := r.db.SelectContext(ctx, &images, query, productID)
	return images, err
}

func (r *productRepository) GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error) {
	var category domain.Category
	query := `SELECT * FROM categories WHERE id = ?`
	err := r.db.GetContext(ctx, &category, query, categoryID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &category, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domain.ProductVariant, error) {
	var variant domain.ProductVariant
	query := `SELECT * FROM product_variants WHERE id = ?`
	err := r.db.GetContext(ctx, &variant, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &variant, nil
}
