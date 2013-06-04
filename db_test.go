package jet

import (
	_ "github.com/bmizerany/pq"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}
	l := NewLogger(os.Stdout)
	db.SetLogger(l)
	db.SetColumnConverter(SnakeCaseConverter)
	if db.Logger() != l {
		t.Fatal("wrong logger set")
	}
	err = db.Query(`DROP TABLE IF EXISTS "table"`).Run()
	if err != nil {
		t.Fatal(err.Error())
	}
	err = db.Query(`CREATE TABLE "table" ( "a" text, "b" integer )`).Run()
	if err != nil {
		t.Fatal(err.Error())
	}
	var mv map[string]interface{}
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello", 7).Rows(&mv)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := len(mv); x != 1 {
		t.Fatal("wrong map len", x, mv)
	}
	x, ok := mv["a"].([]uint8)
	if !ok || string(x) != "hello" {
		t.Fatal(x)
	}
	var sv struct{ A string }
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello2", 8).Rows(&sv)
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
	err = db.Query(`SELECT * FROM "table"`).Rows(&sv2)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := len(sv2); x != 2 {
		t.Fatal(x, sv2)
	}
	if x := sv2[0]; x.A != "hello" || x.B != 7 {
		t.Fatal(x)
	}
	if x := sv2[1]; x.A != "hello2" || x.B != 8 {
		t.Fatal(x)
	}
	var sv3 []data
	err = db.Query(`SELECT * FROM "table"`).Rows(&sv3, 1)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := len(sv3); x != 1 {
		t.Fatal(x)
	}
	if x := sv3[0]; x.A != "hello" || x.B != 7 {
		t.Fatal(x)
	}

	// Test Value()
	var b int64
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "b"`, "hellov", 101).Value(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 101 {
		t.Fatal(b)
	}
}
