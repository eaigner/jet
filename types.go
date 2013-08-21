package jet

import (
	"database/sql"
)

type Migration struct {
	Up   string
	Down string
	Id   int64
}

type Runnable interface {
	// Run runs the query without returning results
	Run() error
	// Rows runs the query writing the rows to the specified map or struct array. If maxRows is specified, only writes up to maxRows rows.
	Rows(v interface{}, maxRows ...int64) error
}

type queryObject interface {
	Prepare(query string) (*sql.Stmt, error)
}
