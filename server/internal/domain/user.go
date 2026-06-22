package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

type NullString struct {
	String string
	Valid  bool
}

func (ns *NullString) Scan(value any) error {
	if value == nil {
		ns.String = ""
		ns.Valid = false
		return nil
	}

	switch v := value.(type) {
	case string:
		ns.String = v
		ns.Valid = true
		return nil
	case []byte:
		ns.String = string(v)
		ns.Valid = true
		return nil
	case sql.RawBytes:
		ns.String = string(v)
		ns.Valid = false
		return nil
	default:
		ns.String = ""
		ns.Valid = false
		return nil
	}
}

func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		ns.String = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ns.String = s
	ns.Valid = true
	return nil
}

func NewNullString(s string) NullString {
	return NullString{
		String: s,
		Valid:  s != "",
	}
}

type User struct {
	ID           int        `db:"id" json:"-"`
	PublicID     string     `db:"public_id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FullName     string     `db:"full_name" json:"full_name"`
	Phone        NullString `db:"phone" json:"phone"`
	Address      NullString `db:"address" json:"address"`
	Role         string     `db:"role" json:"role"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
