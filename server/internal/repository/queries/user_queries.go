package queries

const (
	UserInsert = `
		INSERT INTO users (public_id, email, password_hash, full_name, phone, address, role)
		VALUES (:public_id, :email, :password_hash, :full_name, :phone, :address, :role)`

	UserSelectTimestamps = `SELECT is_active, created_at, updated_at FROM users WHERE id = ?`

	UserSelectByEmail = `SELECT id, public_id, email, password_hash, full_name, phone, address, role, is_active, created_at, updated_at FROM users WHERE email = ?`

	UserSelectByID = `
		SELECT id, public_id, email, password_hash, full_name, phone, address, role, is_active, created_at, updated_at
		FROM users WHERE id = ?`

	UserSelectByPublicID = `
		SELECT id, public_id, email, password_hash, full_name, phone, address, role, is_active, created_at, updated_at
		FROM users WHERE public_id = ?`

	UserUpdate = `UPDATE users SET full_name = :full_name, phone = :phone, address = :address, updated_at = NOW() WHERE id = :id`
)
