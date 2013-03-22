package jet

import (
	"database/sql"
)

type Migration interface {
	Up(tx Tx)
	Down(tx Tx)
}

type MigrationSuite interface {
	Add(m Migration)
	Run(jet Db) error
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
}

type queryObject interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
