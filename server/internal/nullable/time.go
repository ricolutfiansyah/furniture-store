package nullable

import (
	"database/sql"
	"encoding/json"
)

type NullTime struct {
	sql.NullTime
}

// func (nt NullTime) IsZero() bool {
// 	return !nt.Valid
// }

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}
