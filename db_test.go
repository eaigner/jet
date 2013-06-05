package jet

import (
	_ "github.com/bmizerany/pq"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	l := NewLogger(os.Stdout)
	db.SetLogger(l)
	db.SetColumnConverter(SnakeCaseConverter)
	if db.Logger() != l {
		t.Fatal("wrong logger set")
	}
	err = db.Query(`DROP TABLE IF EXISTS "table"`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "table" ( "a" text, "b" integer )`).Run()
	if err != nil {
		t.Fatal(err)
	}

	var mv map[string]interface{}
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello", 7).Rows(&mv)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(mv); x != 1 {
		t.Fatal("wrong map len", x, mv)
	}
	x, ok := mv["a"].([]uint8)
	if !ok || string(x) != "hello" {
		t.Fatal(x)
	}

	var sv struct {
		A string
	}
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "a"`, "hello2", 8).Rows(&sv)
	if err != nil {
		t.Fatal(err)
	}
	if x := sv.A; x != "hello2" {
		t.Fatal(x)
	}

	var sv2 []struct {
		A string
		B int16
	}
	err = db.Query(`SELECT * FROM "table"`).Rows(&sv2)
	if err != nil {
		t.Fatal(err)
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

	var sv3 []struct {
		A string
		B int64
	}
	err = db.Query(`SELECT * FROM "table"`).Rows(&sv3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(sv3); x != 1 {
		t.Fatal(x)
	}
	if x := sv3[0]; x.A != "hello" || x.B != 7 {
		t.Fatal(x)
	}

	// Single value
	var b int64
	err = db.Query(`INSERT INTO "table" ( "a", "b" ) VALUES ( $1, $2 ) RETURNING "b"`, "hellov", 101).Value(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 101 {
		t.Fatal(b)
	}
}

func BenchmarkQueryRows(b *testing.B) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		b.Fatal(err)
	}
	err = db.Query(`DROP TABLE IF EXISTS "benchmark"`).Run()
	if err != nil {
		b.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "benchmark" ( "a" text, "b" integer )`).Run()
	if err != nil {
		b.Fatal(err)
	}
	err = db.Query(`INSERT INTO "benchmark" ( "a", "b" ) VALUES ( $1, $2 )`, "benchme!", 9).Run()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		var v []struct {
			A string
			B int64
		}
		err = db.Query(`SELECT * FROM "table"`).Rows(&v)
		if err != nil {
			b.Fatal(err)
		}
	}
}
