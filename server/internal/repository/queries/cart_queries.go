package queries

const (
	CartInsert = `INSERT INTO carts (user_id) VALUES (?)`

	CartSelectByUserID = `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = ?`

	CartItemSelectByCartID = `
		SELECT id, cart_id, variant_id, quantity, price_at_time, created_at, updated_at
		FROM cart_items
		WHERE cart_id = ?`

	CartItemSelectByUserIDTx = `
		SELECT ci.id, ci.cart_id, ci.variant_id, ci.quantity, ci.price_at_time, ci.created_at
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE c.user_id = ?`

	CartItemInsert = `
		INSERT INTO cart_items (cart_id, variant_id, quantity, price_at_time)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			quantity = quantity + VALUES(quantity)`

	CartItemSelectByCartAndVariant = `
		SELECT id, cart_id, variant_id, quantity, price_at_time, created_at, updated_at
		FROM cart_items
		WHERE cart_id = ? AND variant_id = ?`

	CartItemUpdateQuantity = `
		UPDATE cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		SET ci.quantity = ?
		WHERE ci.id = ? AND c.user_id = ?`

	CartItemDelete = `
		DELETE ci FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE ci.id = ? AND c.user_id = ?`

	// Template for sqlx.In — provide userID and IN clause
	CartItemSelectByIDsTx = `
		SELECT ci.id, ci.cart_id, ci.variant_id, ci.quantity, ci.price_at_time, ci.created_at
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE c.user_id = ? AND ci.id IN (?)
		FOR UPDATE`

	// Template for sqlx.In — provide cartID and IN clause
	CartItemDeleteTx = `DELETE FROM cart_items WHERE cart_id = ? AND id IN (?)`

	// Template for sqlx.In — provide userID and IN clause
	CartItemDeleteBulk = `
		DELETE ci FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE c.user_id = ? AND ci.id IN (?)`

	CartItemSelectByID = `
		SELECT ci.id, ci.cart_id, ci.variant_id, ci.quantity, ci.price_at_time, ci.created_at, ci.updated_at
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE ci.id = ? AND c.user_id = ?`
)
