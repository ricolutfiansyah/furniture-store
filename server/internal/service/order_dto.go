package service

import "furniture-api/internal/nullable"

func toNullString(s string) nullable.NullString {
	if s == "" {
		return nullable.NewNull()
	}

	return nullable.NewNullString(s)
}
