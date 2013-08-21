package jet

import (
	"database/sql"
)

type Tx struct {
	db *Db
	tx *sql.Tx
}

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
