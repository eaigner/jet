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

func TestErrorLogging(t *testing.T) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db.ExpandMapAndSliceMarker(true)
	db.SetColumnConverter(SnakeCaseConverter)
	db.SetLogger(NewLogger(nil))
	err = db.Query(`DROP TABLE IF EXISTS "logtest"`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "logtest" ( "id" serial PRIMARY KEY , "text" text )`).Run()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	var results []struct {
		Id   int64
		Text string
	}
	err = db.Query(`SELECT * FROM "logtest" WHERE id = $1`, 1234).Rows(&results)
	if err != nil { // err should be nil since it was a valid query.
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatal("no results should be found.")
	}

	err = db.Query(`SELECT * FROM "logtest" WHERE id = $1`, "hello").Rows(&results)
	if err == nil { // this query will procude an error.
		t.Fatal("This should produce an error.")
	}

	err = db.Query(`SELECT * FROM "logtest" WHERE id = $1`, 5678).Rows(&results)
	if err != nil { // r.lastErr should be clear and no error should occur.
		t.Fatal("This should produce an error.")
	}
	if len(results) != 0 {
		t.Fatal("no results should be found.")
	}
}
