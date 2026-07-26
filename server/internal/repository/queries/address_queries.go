package queries

const (
	AddressInsert = `
		INSERT INTO user_addresses
			(user_id, label, recipient_name, phone, province, city, district, postal_code, address_line, is_default)
		VALUES
			(:user_id, :label, :recipient_name, :phone, :province, :city, :district, :postal_code, :address_line, :is_default)`

	AddressCountByUserID = `SELECT COUNT(*) FROM user_addresses WHERE user_id = ?`

	AddressSelectByID = `
		SELECT id, user_id, label, recipient_name, phone, province, city, district,
				postal_code, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE id = ? AND user_id = ?`

	AddressSelectByIDForUpdate = `
		SELECT id, user_id, label, recipient_name, phone, province, city, district,
				postal_code, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE id = ? AND user_id = ?
		FOR UPDATE`

	AddressSelectByUserID = `
		SELECT id, user_id, label, recipient_name, phone, province, city, district,
				postal_code, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = ?
		ORDER BY is_default DESC, created_at DESC`

	AddressSelectByUserIDTx = `
		SELECT id, user_id, label, recipient_name, phone, province, city, district,
				postal_code, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = ?`

	AddressUpdate = `
		UPDATE user_addresses
		SET label = :label, recipient_name = :recipient_name, phone = :phone,
			province = :province, city = :city, district = :district,
			postal_code = :postal_code, address_line = :address_line
		WHERE id = :id AND user_id = :user_id`

	AddressDelete = `DELETE FROM user_addresses WHERE id = ? AND user_id = ?`

	AddressUnsetDefault = `UPDATE user_addresses SET is_default = FALSE WHERE user_id = ?`

	AddressSetDefault = `UPDATE user_addresses SET is_default = TRUE WHERE id = ? AND user_id = ?`
)
