package jet

import (
	"database/sql"
)

type tx struct {
	db *sql.DB
}

func (t *tx) Commit() error {
	return nil
}

func (t *tx) Query(sql string, args ...interface{}) error {
	return nil
}
