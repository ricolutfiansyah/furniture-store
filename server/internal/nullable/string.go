package nullable

import (
	"bytes"
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}

func NewNullString(s string) NullString {
	if s == "" {
		return NullString{
			NullString: sql.NullString{
				Valid: false,
			},
		}
	}
	return NullString{
		NullString: sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

func NewNull() NullString {
	return NullString{
		NullString: sql.NullString{
			Valid: false,
		},
	}
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
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
