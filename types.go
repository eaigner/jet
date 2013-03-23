package jet

import (
	"database/sql"
)

type Migration struct {
	Up   string
	Down string
	Id   int64
}

type Db interface {
	Queryable
	// Begin starts a transaction
	Begin() (Tx, error)
}

type Tx interface {
	Queryable
	// Commit commits the transaction
	Commit() error
	// Errors returns all errors that occurred during the transaction
	Errors() []error
}

type Queryable interface {
	// Query prepares the query for execution
	Query(query string, args ...interface{}) Queryable
	// Run runs the query without returning results
	Run() error
	// Rows runs the query writing the rows to the specified map or struct array. If maxRows is specified, only writes up to maxRows rows.
	Rows(v interface{}, maxRows ...int64) error
	// Value returns the value of the first row and column returned. This is useful for e.g. fetching aggregated values.
	Value() (interface{}, error)
	// Logger returns the current logger
	Logger() *Logger
	// SetLogger sets a logger
	SetLogger(l *Logger)
}

type queryObject interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
