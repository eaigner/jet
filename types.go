package jet

import (
	"database/sql"
)

type Runnable interface {
	// Run runs the query returning the sql.Result
	Exec() (sql.Result, error)

	// Run runs the query without returning results
	Run() error
	// Rows runs the query writing the rows to the specified map or struct array.
	// If maxRows is specified, only writes up to maxRows rows.
	Rows(v interface{}, maxRows ...int64) error
}

type queryObject interface {
	Prepare(query string) (*sql.Stmt, error)
}
