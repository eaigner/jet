package jet

import (
	"database/sql"
)

// Tx represents a transaction instance.
// It can be created using Begin on the *Db object.
type Tx struct {
	db  *Db
	tx  *sql.Tx
	qid string
}

// Query creates a prepared query that can be run with Rows or Run.
func (tx *Tx) Query(query string, args ...interface{}) Runnable {
	q := newQuery(tx.tx, tx.db, query, args...)
	q.id = tx.qid
	return q
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	if tx.db.LogFunc != nil {
		tx.db.LogFunc(tx.qid, "COMMIT")
	}
	return tx.tx.Commit()
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	if tx.db.LogFunc != nil {
		tx.db.LogFunc(tx.qid, "ROLLBACK")
	}
	return tx.tx.Rollback()
}
