package jet

import (
	"context"
	"database/sql"
)

type Runnable interface {
	// Run runs the query without returning results
	Run() error
	// Rows runs the query writing the rows to the specified map or struct array.
	Rows(v interface{}) error
}

type queryObject interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// ComplexValue implements methods for en/decoding custom values to a format the driver understands.
type ComplexValue interface {
	Encode() interface{}

	// Decode receives a plain value to decode, never a pointer.
	Decode(v interface{}) error
}
