package jet

import (
	"database/sql"
)

type tx struct {
	tx     *sql.Tx
	runner *runner
	errors []error
}

func (t *tx) Commit() error {
	panic("not implemented")
}

func (t *tx) Query(query string, args ...interface{}) Queryable {
	t.runner = &runner{
		qo:    t.tx,
		query: query,
		args:  args,
	}
	return t
}

func (t *tx) Run() error {
	return t.runner.Run()
}

func (t *tx) Rows(v interface{}, maxRows ...int64) error {
	return t.runner.Rows(v, maxRows...)
}
