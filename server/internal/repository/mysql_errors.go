package repository

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func isDuplicateKeyError(err error, key string) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return strings.Contains(mysqlErr.Message, key)
	}

	return false
}
