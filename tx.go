package jet

import (
	"database/sql"
)

type tx struct {
	runner
	tx     *sql.Tx
	id     string
	errors []error
}

func (t *tx) Commit() error {
	if l := t.Logger(); l != nil {
		l.Actionf("COMMIT TRANSACTON %s", t.id)
	}
	return t.appendError(t.tx.Commit())
}

func (t *tx) Query(query string, args ...interface{}) Queryable {
	t.runner = runner{
		qo:     t.tx,
		query:  query,
		args:   args,
		logger: t.runner.logger,
	}
	return t
}

func (t *tx) Run() error {
	return t.appendError(t.runner.Run())
}

func (t *tx) Rows(v interface{}, maxRows ...int64) error {
	return t.appendError(t.runner.Rows(v, maxRows...))
}

func (t *tx) Errors() []error {
	return t.errors
}

func (t *tx) appendError(err error) error {
	if err != nil {
		if t.errors == nil {
			t.errors = []error{err}
		} else {
			t.errors = append(t.errors, err)
		}
	}
	return err
}
