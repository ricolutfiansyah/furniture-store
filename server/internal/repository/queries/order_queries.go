package queries

const (
	OrderInsert = `
		INSERT INTO orders
			(user_id, order_number, total_amount, shipping_cost, tax, grand_total,
			status, shipping_address, payment_method, notes
		) VALUES (:user_id, :order_number, :total_amount, :shipping_cost, :tax, :grand_total,
			:status, :shipping_address, :payment_method, :notes)`

	OrderSelectTimestamps = `SELECT created_at, updated_at FROM orders WHERE id = ?`

	OrderItemInsert = `
		INSERT INTO order_items (order_id, variant_id, quantity, price_per_item, total_price)
		VALUES (:order_id, :variant_id, :quantity, :price_per_item, :total_price)`

	OrderItemSelectTimestamp = `SELECT created_at FROM order_items WHERE id = ?`

	OrderStatusInsert = `
		INSERT INTO order_statuses (order_id, status, notes, created_by)
		VALUES (?, ?, ?, ?)`

	OrderSelectByUserID = `
		SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
			status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
			notes, created_at, updated_at
		FROM orders WHERE user_id = ?
		ORDER BY created_at DESC`

	OrderSelectByID = `
		SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
			status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
			notes, created_at, updated_at
		FROM orders WHERE id = ? AND user_id = ?`

	OrderSelectByIDForAdmin = `
		SELECT id, user_id, order_number, total_amount, shipping_cost, tax, grand_total,
			status, shipping_address, payment_method, paid_at, shipped_at, delivered_at,
			notes, created_at, updated_at
		FROM orders WHERE id = ?`

	OrderItemSelectByOrderID = `
		SELECT id, order_id, variant_id, quantity, price_per_item, total_price, created_at
		FROM order_items WHERE order_id = ?`

	OrderStatusSelectByOrderID = `
		SELECT id, order_id, status, notes, created_by, created_at
		FROM order_statuses WHERE order_id = ? ORDER BY created_at ASC`

	OrderStatusUpdate = `UPDATE orders SET status = ?`

	OrderStatusSelectForUpdate = `SELECT status FROM orders WHERE id = ? FOR UPDATE`

	OrderSummarySelect = `
		SELECT oi.order_id, (SELECT COUNT(*) FROM order_items WHERE order_id = oi.order_id) as total_items, pv.variant_name,
			COALESCE(pi.image_url, '') as image_url
		FROM order_items oi
		JOIN product_variants pv ON oi.variant_id = pv.id
		LEFT JOIN product_images pi ON pv.product_id = pi.product_id AND pi.is_primary = TRUE
		WHERE oi.id IN (SELECT MIN(id) FROM order_items WHERE order_id IN (?) GROUP BY order_id)`
)
