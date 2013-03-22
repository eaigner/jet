package jet

import (
	"database/sql"
)

type tx struct {
	godb *sql.DB
}

func (t *tx) Commit() error {
	return nil
}

func (t *tx) Query(query string, args ...interface{}) Queryable {
	panic("not implemented")
}

func (t *tx) Run() error {
	panic("not implemented")
}

func (t *tx) Rows(v interface{}, maxRows ...int64) error {
	panic("not implemented")
}
