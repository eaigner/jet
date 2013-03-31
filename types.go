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
	// Rollback rolls back the transaction
	Rollback() error
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
	// Value writes the value of the first row and column to the passed pointer. If the value cannot be converted to the specified type, the method will panic. This is useful for e.g. fetching aggregated values.
	Value(v interface{}) error
	// Logger returns the current logger
	Logger() *Logger
	// SetLogger sets a logger
	SetLogger(l *Logger)
}

type queryObject interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
