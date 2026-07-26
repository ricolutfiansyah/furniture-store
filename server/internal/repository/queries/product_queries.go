package queries

const (
	ProductSelectActive = `
		SELECT id, category_id, name, slug, description, base_price, sku,
		weight_kg, is_active, views, created_at, updated_at
		FROM products
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	ProductCountActive = `SELECT COUNT(*) FROM products WHERE is_active = TRUE`

	ProductSelectBySlug = `
		SELECT id, category_id, name, slug, description, base_price, sku,
		weight_kg, is_active, views, created_at, updated_at
		FROM products
		WHERE slug = ? AND is_active = TRUE`

	ProductSelectByID = `
		SELECT id, category_id, name, slug, description, base_price, sku,
		weight_kg, is_active, views, created_at, updated_at
		FROM products
		WHERE id = ?`

	VariantSelectByProductID = `
		SELECT id, product_id, variant_name, attributes, additional_price, stock_quantity, sku_variant,
		weight_kg, is_active, created_at, updated_at
		FROM product_variants
		WHERE product_id = ? AND is_active = TRUE`

	ImageSelectByProductID = `
		SELECT id, product_id, variant_id, image_url, is_primary, sort_order, created_at
		FROM product_images
		WHERE product_id = ?
		ORDER BY is_primary DESC, sort_order ASC`

	CategorySelectByID = `SELECT id, name, slug, description, parent_id, image_url, created_at, updated_at FROM categories WHERE id = ?`

	VariantSelectByID = `
		SELECT id, product_id, variant_name, attributes, additional_price, stock_quantity,
		sku_variant, weight_kg, is_active, created_at, updated_at
		FROM product_variants
		WHERE id = ?`

	VariantDecreaseStock = `UPDATE product_variants SET stock_quantity = stock_quantity - ? WHERE id = ? AND stock_quantity >= ?`
)
