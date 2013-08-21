package jet

import (
	"database/sql"
)

// Tx represents a transaction instance.
// It can be created using Begin on the *Db object.
type Tx struct {
	db *Db
	tx *sql.Tx
}

// Query creates a prepared query that can be run with Rows or Run.
func (t *Tx) Query(query string, args ...interface{}) Runnable {
	return newQuery(t.tx, t.db).prepare(query, args...)
}

// Commit commits the transaction
func (t *Tx) Commit() error {
	// if l := t.Logger(); l != nil {
	// 	l.Txnf("COMMIT   %s", t.txnId).Println()
	// }
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Tx) Rollback() error {
	// if l := t.Logger(); l != nil {
	// 	l.Txnf("ROLLBACK %s", t.txnId).Println()
	// }
	return t.tx.Rollback()
}
