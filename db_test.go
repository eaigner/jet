package jet

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	drivers := map[string]string{
		"postgres": "user=postgres dbname=jet sslmode=disable",
		"mysql":    "benchmarkdbuser:benchmarkdbpass@tcp(localhost:3306)/hello_world?charset=utf8", // TODO: change source
	}
	for driverName, driverSource := range drivers {
		t.Log(driverName)

		db, err := Open(driverName, driverSource)
		if err != nil {
			t.Fatal(err)
		}
		l := NewLogger(os.Stdout)
		db.SetLogger(l)
		db.SetColumnConverter(SnakeCaseConverter)
		if db.Logger() != l {
			t.Fatal("wrong logger set")
		}
		err = db.Query(`DROP TABLE IF EXISTS jetTest`).Run()
		if err != nil {
			t.Fatal(err)
		}
		err = db.Query(`CREATE TABLE jetTest ( a VARCHAR(100), b INT )`).Run()
		if err != nil {
			t.Fatal(err)
		}

		t.Log("map unpack")

		var mv map[string]interface{}
		switch driverName {
		case "postgres":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING a`, "hello", 7).Rows(&mv)
		case "mysql":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( ?, ? )`, "hello", 7).Run()
		}
		if err != nil {
			t.Fatal(err)
		}
		if driverName == "postgres" {
			if x := len(mv); x != 1 {
				t.Fatal("wrong map len", x, mv)
			}
			x, ok := mv["a"].([]uint8)
			if !ok || string(x) != "hello" {
				t.Fatal(x)
			}
		}

		t.Log("struct unpack")

		var sv struct {
			A string
		}
		switch driverName {
		case "postgres":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING a`, "hello2", 8).Rows(&sv)
		case "mysql":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( ?, ? )`, "hello2", 8).Run()
		}
		if err != nil {
			t.Fatal(err)
		}
		if driverName == "postgres" {
			if x := sv.A; x != "hello2" {
				t.Fatal(x)
			}
		}

		t.Log("struct slice unpack")

		var sv2 []struct {
			A string
			B int16
		}
		err = db.Query(`SELECT * FROM jetTest`).Rows(&sv2)
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

		t.Log("struct slice unpack 2")

		var sv3 []struct {
			A string
			B int64
		}
		err = db.Query(`SELECT * FROM jetTest`).Rows(&sv3, 1)
		if err != nil {
			t.Fatal(err)
		}
		if x := len(sv3); x != 1 {
			t.Fatal(x)
		}
		if x := sv3[0]; x.A != "hello" || x.B != 7 {
			t.Fatal(x)
		}

		t.Log("single value")

		var b int64
		switch driverName {
		case "postgres":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING b`, "hellov", 101).Value(&b)
		case "mysql":
			err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( ?, ? )`, "hellov", 101).Run()
		}
		if err != nil {
			t.Fatal(err)
		}
		if driverName == "postgres" {
			if b != 101 {
				t.Fatal(b)
			}
		}
	}
}

func TestNull(t *testing.T) {
	drivers := map[string]string{
		"postgres": "user=postgres dbname=jet sslmode=disable",
	}
	for driverName, driverSource := range drivers {
		t.Log(driverName)

		db, err := Open(driverName, driverSource)
		if err != nil {
			t.Fatal(err)
		}
		db.SetLogger(NewLogger(os.Stdout))
		db.SetColumnConverter(SnakeCaseConverter)

		err = db.Query(`DROP TABLE IF EXISTS jetNullTest`).Run()
		if err != nil {
			t.Fatal(err)
		}
		err = db.Query(`CREATE TABLE jetNullTest ( a VARCHAR(100) )`).Run()
		if err != nil {
			t.Fatal(err)
		}
		err = db.Query(`INSERT INTO jetNullTest ( a ) VALUES ( NULL )`).Run()
		if err != nil {
			t.Fatal(err)
		}

		var rows []struct {
			A string
		}
		err = db.Query(`SELECT * FROM jetNullTest`).Rows(&rows)
		if err != nil {
			t.Fatal(err)
		}
		if x := len(rows); x != 1 {
			t.Fatal(x)
		}
		if x := rows[0].A; x != "" {
			t.Fatal(x)
		}
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v []struct {
			A string
			B int64
		}
		err = db.Query(`SELECT * FROM "benchmark"`).Rows(&v)
		if err != nil {
			b.Fatal(err)
		}
	}
}
