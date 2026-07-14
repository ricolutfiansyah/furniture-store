package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"furniture-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetActive(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	const query = `
		SELECT id, category_id, name, slug, description, base_price, sku, 
		weight_kg, is_active, views, created_at, updated_at 
		FROM products 
		WHERE is_active = TRUE 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`

	products := []domain.Product{}
	err := r.db.SelectContext(ctx, &products, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get active products: %w", err)
	}

	return products, nil
}

func (r *productRepository) CountActive(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM products WHERE is_active = TRUE`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("count active products: %w", err)
	}

	return count, nil
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	query := `
			SELECT id, category_id, name, slug, description, base_price, sku, 
			weight_kg, is_active, views, created_at, updated_at 
			FROM products
			WHERE slug = ? AND is_active = TRUE`

	var product domain.Product
	err := r.db.GetContext(ctx, &product, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by slug: %w", err)
	}

	return &product, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	query := `
		SELECT id, category_id, name, slug, description, base_price, sku,
		weight_kg, is_active, views, created_at, updated_at
		FROM products
		WHERE id = ?`

	var product domain.Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by id: %w", err)
	}

	return &product, nil
}

func (r *productRepository) GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error) {
	const query = `
				SELECT id, product_id, variant_name, attributes, additional_price, stock_quantity, sku_variant, 
				weight_kg, is_active, created_at, updated_at 
				FROM product_variants 
				WHERE product_id = ? AND is_active = TRUE`

	variants := []domain.ProductVariant{}
	err := r.db.SelectContext(ctx, &variants, query, productID)
	if err != nil {
		return nil, fmt.Errorf("get variants by product id: %w", err)
	}

	return variants, nil
}

func (r *productRepository) GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	const query = `
				SELECT id, product_id, variant_id, image_url, is_primary, sort_order, created_at 
				FROM product_images 
				WHERE product_id = ? 
				ORDER_BY is_active DESC, sort_order ASC`

	images := []domain.ProductImage{}
	err := r.db.SelectContext(ctx, &images, query, productID)
	if err != nil {
		return nil, fmt.Errorf("get images by product id: %w", err)
	}

	return images, nil
}

func (r *productRepository) GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error) {
	query := `SELECT id, name, slug, description, parent_id, image_url, created_at, updated_at FROM categories WHERE id = ?`

	var category domain.Category
	err := r.db.GetContext(ctx, &category, query, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}

	return &category, nil
}

func (r *productRepository) GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error) {
	const query = `
				SELECT id, product_id, variant_name, attributes, additional_price, stock_quantity, 
				sku_variant, weight_kg, is_active, created_at, updated_at 
				FROM product_variants 
				WHERE id = ?`

	var variant domain.ProductVariant
	err := r.db.GetContext(ctx, &variant, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVariantNotFound
		}
		return nil, fmt.Errorf("get variant by id: %w", err)
	}
	return &variant, nil
}

func (r *productRepository) DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error {
	query := `UPDATE product_variants SET stock_quantity = stock_quantity - ? WHERE id = ? AND stock_quantity >= ?`

	result, err := tx.ExecContext(ctx, query, quantity, variantID, quantity)
	if err != nil {
		return fmt.Errorf("decrease stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("decrease stock rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInsufficientStock
	}

	return nil
}
