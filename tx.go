package jet

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

// Tx represents a transaction instance.
// It can be created using Begin on the *Db object.
type Tx struct {
	db  *Db
	tx  *sqlx.Tx
	qid string
}

// Query creates a prepared query that can be run with Rows or Run.
func (tx *Tx) Query(query string, args ...interface{}) Runnable {
	return tx.QueryContext(context.Background(), query, args...)
}

// QueryContext creates a prepared query that can be run with Rows or Run.
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) Runnable {
	q := newQuery(ctx, tx.tx, tx.db, query, args...)
	q.id = tx.qid
	return q
}

// Exec calls Exec on the underlying sql.Tx.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	if tx == nil || tx.tx == nil {
		return nil, errors.New("jet: Exec called on nil transaction")
	}
	return tx.tx.Exec(query, args...)
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	tx.db.LogFunc(context.Background(), tx.qid, "COMMIT")
	return tx.tx.Commit()
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	tx.db.LogFunc(context.Background(), tx.qid, "ROLLBACK")
	return tx.tx.Rollback()
}
