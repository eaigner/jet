package jet

import (
	"testing"
)

func TestQueryMarkSubstitution(t *testing.T) {
	query := `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, $2, $3 )`
	args := []interface{}{"banana", map[string]interface{}{
		"color": "yellow",
		"price": 2,
	}, "cuba"}

	// Maps
	newquery, newargs := substituteMapAndArrayMarks(query, args...)
	if newquery != `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, hstore(ARRAY[ $2, $3, $4, $5 ]), $6 )` {
		t.Fatal(newquery)
	}
	if x := len(newargs); x != 6 {
		t.Fatal(x)
	}

	// Slice
	query2 := `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, ARRAY[ $2 ], $3 )`
	args2 := []interface{}{"banana", []interface{}{"a", "b", "c"}, "cuba"}

	t.Log(query)
	t.Log(args2)

	newquery, newargs = substituteMapAndArrayMarks(query2, args2...)
	if newquery != `INSERT INTO "fruits" ( "name", "attrs", "origin" ) VALUES ( $1, ARRAY[ $2, $3, $4 ], $5 )` {
		t.Fatal(newquery)
	}
	if x := len(newargs); x != 5 {
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
