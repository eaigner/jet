package jet

import (
	"testing"
)

func TestHstoreQueryMarkSubstitution(t *testing.T) {
	query := `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, $2, $3 )`
	args := []interface{}{"banana", map[string]interface{}{
		"color": "yellow",
		"price": 2,
	}, "cuba"}

	newquery, newargs := substituteHstoreMarks(query, args...)
	if newquery != `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, hstore(ARRAY[$2, $3, $4, $5]), $6 )` {
		t.Fatal(newquery)
	}
	if x := len(newargs); x != 6 {
		t.Fatal(x)
	}
}

func TestHstoreColumnParse(t *testing.T) {
	h := parseHstoreColumn(`"a\"b\"c"=>"d\"e\"f", "g\"h\"i"=>"j\"k\"l"`)
	if x := len(h); x != 2 {
		t.Fatal(h)
	}
	t.Log(h)
	if x := h[`a"b"c`]; x != `d"e"f` {
		t.Fatal(x)
	}
}

func TestHstoreQuery(t *testing.T) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db.SetLogger(NewLogger(nil))
	err = db.Query(`CREATE EXTENSION IF NOT EXISTS hstore`).Run()
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
		`INSERT INTO "hstoretable" VALUES ( $1, $2, $3 )`,
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
