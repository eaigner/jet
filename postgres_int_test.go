package jet

// This file contains all integration tests for PostgreSQL.
// All tests have to be prefixed with (Test|Benchmark)IntPg.

import (
	_ "github.com/lib/pq"
	"testing"
	"time"
)

func openPg(t *testing.T) *Db {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	db.LogFunc = func(queryId, query string, args ...interface{}) {
		t.Logf("%s: %s ARG: %v", queryId, query, args)
	}

	err = db.Query("DROP SCHEMA public CASCADE").Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query("CREATE SCHEMA public").Run()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type cx struct {
	a string
	b string
}

func (c *cx) Encode() interface{} {
	return c.a + c.b
}

func (c *cx) Decode(v interface{}) error {
	var s string
	switch t := v.(type) {
	case []uint8:
		s = string(t)
	case string:
		s = t
	}
	c.a = string(s[0])
	c.b = string(s[1])

	return nil
}

func TestComplexValues(t *testing.T) {
	db := openPg(t)

	err := db.Query(`CREATE TABLE complexTest ( a text )`).Run()
	if err != nil {
		t.Fatal(err)
	}

	var c cx
	err = db.Query(`INSERT INTO complexTest ( a ) VALUES ( $1 ) RETURNING a`, &cx{"x", "y"}).Rows(&c)
	if err != nil {
		t.Fatal(err)
	}

	if c.a != "x" || c.b != "y" {
		t.Fatal(c)
	}
}

func TestPgRowUnpack(t *testing.T) {
	db := openPg(t)
	err := db.Query(`DROP TABLE IF EXISTS jetTest`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE jetTest ( a VARCHAR(100), b INT )`).Run()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("map unpack")

	var mv map[string]interface{}
	err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING a`, "hello", 7).Rows(&mv)
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

	t.Log("struct unpack")

	var sv struct {
		A string
	}
	err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING a`, "hello2", 8).Rows(&sv)
	if err != nil {
		t.Fatal(err)
	}
	if x := sv.A; x != "hello2" {
		t.Fatal(x)
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

	t.Log("struct slice unpack with limit")

	var sv3 []struct {
		A string
		B int64
	}
	err = db.Query(`SELECT * FROM jetTest`).Rows(&sv3)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(sv3); x != 2 {
		t.Fatal(x)
	}
	if x := sv3[0]; x.A != "hello" || x.B != 7 {
		t.Fatal(x)
	}

	t.Log("single value")

	var b int64
	err = db.Query(`INSERT INTO jetTest ( a, b ) VALUES ( $1, $2 ) RETURNING b`, "hellov", 101).Rows(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 101 {
		t.Fatal(b)
	}
}

func TestPgTransaction(t *testing.T) {
	db := openPg(t)
	err := db.Query(`DROP TABLE IF EXISTS "tx_table"`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "tx_table" ( "a" text, "b" integer )`).Run()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err1 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello", 7).Run()
	if err1 != nil {
		t.Fatal(err1.Error())
	}
	err2 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello2", time.Now()).Run()
	if err2 == nil {
		t.Fatal("should return error")
	}
	err3 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello2", "boo").Run()
	if err3 == nil {
		t.Fatal("should return error")
	}

	// Commit now returns errors
	// https://github.com/lib/pq/commit/f8ffc32df8b9c5fd7d5ca1ac8345d75e82234edd
	err = tx.Commit()
	if err == nil {
		t.Fatal(err)
	}

	// Rollback
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "roll-me-back", 14).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}
	var c int64
	err = db.Query(`SELECT COUNT(*) FROM "tx_table"`).Rows(&c)
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatal(c)
	}
}

func TestPgNullValue(t *testing.T) {
	db := openPg(t)

	err := db.Query(`DROP TABLE IF EXISTS jetNullTest`).Run()
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

func TestPgHstoreQuery(t *testing.T) {
	db := openPg(t)
	err := db.Query(`CREATE EXTENSION IF NOT EXISTS hstore`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`DROP TABLE IF EXISTS "hstoretable"`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "hstoretable" ( "a" text, "b" hstore, "c" text )`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(
		`INSERT INTO "hstoretable" VALUES ( $1, hstore(ARRAY[ $2 ]), $3 )`,
		"aval",
		map[string]interface{}{"key1": "val1", "key2": 2},
		"cval",
	).Run()
	if err != nil {
		t.Fatal(err)
	}
	var results []struct {
		A string
		B map[string]interface{}
		C string
	}
	err = db.Query(`SELECT * FROM "hstoretable"`).Rows(&results)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(results); x != 1 {
		t.Fatal(x)
	}
	r := results[0]
	if r.A != "aval" {
		t.Fatal(r)
	}
	if r.B["key1"] != "val1" {
		t.Fatal(r)
	}
	if r.B["key2"] != "2" {
		t.Fatal(r)
	}
	if r.C != "cval" {
		t.Fatal(r)
	}
}

func TestPgErrors(t *testing.T) {
	db := openPg(t)
	err := db.Query(`DROP TABLE IF EXISTS "logtest"`).Run()
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

func TestPgUniqueIndex(t *testing.T) {
	db := openPg(t)

	run := func(query string) {
		_, err := db.Exec(query)
		if err != nil {
			t.Fatal(err)
		}
	}

	run(`DROP TABLE IF EXISTS "unique"`)
	run(`DROP INDEX IF EXISTS "unique_field_ix"`)
	run(`CREATE TABLE IF NOT EXISTS "unique" ( "field" text )`)
	run(`CREATE UNIQUE INDEX "unique_field_ix" ON "unique" ( "field" );`)

	for i := 0; i < 2; i++ {
		err := db.Query(`INSERT INTO "unique" ( "field" ) VALUES ( $1 )`, "banana").Run()
		switch i {
		case 1:
			if err == nil {
				t.Fatal(err)
			}
		default:
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestPgSuite(t *testing.T) {
	var s Suite
	s.AddSQL(
		`CREATE TABLE "suite_test" ( id serial, name text )`,
		`DROP TABLE "suite_test"`,
	)
	s.AddSQL(
		`ALTER TABLE "suite_test" ADD "newcol" integer`,
		`ALTER TABLE "suite_test" DROP COLUMN "newcol"`,
	)
	s.AddSQL(
		`ALTER TABLE "suite_test" ADD "anothercol" bytea`,
		`ALTER TABLE "suite_test" DROP COLUMN "anothercol"`,
	)
	s.AddSQL(
		`CREATE INDEX "name_index" ON "suite_test" ( "name" )`,
		`DROP INDEX "name_index"`,
	)

	db := openPg(t)

	if c, s, err := s.Run(db, true, 0); err != nil || c != 4 || s != 4 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Run(db, false, 0); err != nil || c != 0 || s != 4 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Migrate(db); err != nil || c != 4 || s != 4 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Reset(db); err != nil || c != 0 || s != 4 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Step(db); err != nil || c != 1 || s != 1 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Step(db); err != nil || c != 2 || s != 1 {
		t.Fatal(c, s, err)
	}
	if c, s, err := s.Rollback(db); err != nil || c != 1 || s != 1 {
		t.Fatal(c, s, err)
	}
}

func TestPgSuiteError(t *testing.T) {
	var s Suite
	s.AddSQL(`CREATE TABLE flowers (id serial)`, `DROP TABLE flowers`)
	s.AddSQL(`CREATE TABLE err ors (id serial)`, `DROP TABLE errors`)

	db := openPg(t)
	cur, applied, err := s.Migrate(db)

	if err == nil {
		t.Fatal("expected error")
	} else {
		t.Log(err)
	}
	if cur != 1 {
		t.Fatal(cur)
	}
	if applied != 1 {
		t.Fatal(applied)
	}
}

func Benchmark_PgQuery(b *testing.B) {
	db := openPg(nil)
	err := db.Query(`DROP TABLE IF EXISTS "benchmark"`).Run()
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
