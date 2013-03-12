package jet

import (
	_ "github.com/bmizerany/pq"
	"testing"
)

func TestQuery(t *testing.T) {
	db, err := Open("postgres", "user=jet dbname=jet sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = db.Exec(`DROP TABLE IF EXISTS "table"`)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = db.Exec(`CREATE TABLE "table" ( "a" text, "b" integer )`)
	if err != nil {
		t.Fatal(err.Error())
	}
	var mv map[string]interface{}
	err = db.Query(&mv, `INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello", 7)
	if err != nil {
		t.Fatal(err.Error())
	}
	x, ok := mv["a"].([]uint8)
	if !ok || string(x) != "hello" {
		t.Fatal(x)
	}
	var sv struct{ A string }
	err = db.Query(&sv, `INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello2", 8)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := sv.A; x != "hello2" {
		t.Fatal(x)
	}
	type data struct {
		A string
		B int64
	}
	var sv2 []data
	err = db.Query(&sv2, `SELECT * FROM "table"`)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := len(sv2); x != 2 {
		t.Fatal(x)
	}
	if x := sv2[0]; x.A != "hello" || x.B != 7 {
		t.Fatal(x)
	}
	if x := sv2[1]; x.A != "hello2" || x.B != 8 {
		t.Fatal(x)
	}
}
