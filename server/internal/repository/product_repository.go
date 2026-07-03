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

	var products []domain.Product
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
	query := `SELECT id, category_id, name, slug, description, base_price, sku, 
				weight_kg, is_active, views, created_at, updated_at, 
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

func (r *productRepository) GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error) {
	var variants []domain.ProductVariant
	query := `SELECT * FROM product_variants WHERE product_id = ? AND is_active = TRUE`
	err := r.db.SelectContext(ctx, &variants, query, productID)
	return variants, err
}

func (r *productRepository) GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error) {
	const query = `SELECT id, product_id, variant_id, image_url, is_primary, sort_order, created_at 
					FROM product_images 
					WHERE product_id = ? 
					ORDER_BY is_active DESC, sort_order ASC`

	var images []domain.ProductImage
	err := r.db.SelectContext(ctx, &images, query, productID)
	if err != nil {
		return nil, fmt.Errorf("get images by product id: %w", err)
	}

	return images, nil
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

func (r *productRepository) DecreaseStockWithTx(ctx context.Context, tx *sqlx.Tx, variantID, quantity int) error {
	query := `UPDATE product_variants SET stock_quantity = stock_quantity - ? WHERE id = ? AND stock_quantity >= ?`
	result, err := tx.ExecContext(ctx, query, quantity, variantID, quantity)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("Insufficient stock")
	}

	return nil
}
