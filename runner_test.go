package jet

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestCopy(t *testing.T) {
	r := &runner{
		db:     new(sql.DB),
		tx:     new(sql.Tx),
		stmt:   new(sql.Stmt),
		qo:     new(sql.DB),
		conv:   SnakeCaseConverter,
		expand: true,
		txnId:  "A",
		query:  "B",
		args:   []interface{}{"C"},
		errors: []error{},
		logger: NewLogger(nil),
	}
	r2 := r.copy()
	if r == r2 {
		t.Fail()
	}
	if !reflect.DeepEqual(r, r2) {
		t.Log(r)
		t.Log(r2)
		t.Fail()
	}
}
