package jet

type tx struct {
	runner
}

func (t *tx) Commit() error {
	// if l := t.Logger(); l != nil {
	// 	l.Txnf("COMMIT   %s", t.txnId).Println()
	// }
	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	// if l := t.Logger(); l != nil {
	// 	l.Txnf("ROLLBACK %s", t.txnId).Println()
	// }
	return t.tx.Rollback()
}
