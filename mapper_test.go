package jet

import (
	"reflect"
	"testing"
)

func TestColumnToFieldName(t *testing.T) {
	if x := columnToFieldName("column_one"); x != "ColumnOne" {
		t.Fatal(x)
	}
	if x := columnToFieldName("a"); x != "A" {
		t.Fatal(x)
	}
	if x := columnToFieldName("ab"); x != "Ab" {
		t.Fatal(x)
	}
	if x := columnToFieldName("Already_Camel"); x != "AlreadyCamel" {
		t.Fatal(x)
	}
}

func TestUnpackStruct(t *testing.T) {
	m := map[string]interface{}{
		"ab_c": int64(9),
		"c_d":  "hello",
		"e":    "unsettable",
		"f":    []uint8("uint8str"),
		"g":    []uint8("uint8data"),
	}
	type out struct {
		AbC int64
		CD  string
		e   string
		F   string
		G   []byte
	}

	// Unpack struct
	var v out
	err := mapper{m}.unpack(v)
	if err == nil {
		t.Fatal("should return error")
	}
	err = mapper{m}.unpack(&v)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := v.AbC; x != 9 {
		t.Fatal(x)
	}
	if x := v.CD; x != "hello" {
		t.Fatal(x)
	}
	if x := v.e; x != "" {
		t.Fatal(x)
	}
	if x := v.F; x != "uint8str" {
		t.Fatal(x)
	}
	if x := v.G; string(x) != "uint8data" {
		t.Fatal(x)
	}
}

func TestUnpackMap(t *testing.T) {
	m := map[string]interface{}{
		"ab_c": int64(9),
		"c_d":  "hello",
		"e":    "unsettable",
	}
	type out struct {
		AbC int64
		CD  string
		e   string
	}
	var m2 map[string]interface{}
	err := mapper{m}.unpack(&m2)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(m, m2) {
		t.Fatalf("%v\n\n%v\n", m, m2)
	}
}

func TestUnpackStructSlice(t *testing.T) {
	m := map[string]interface{}{
		"A": int64(1),
		"B": "hello",
	}
	m2 := map[string]interface{}{
		"A": int64(2),
		"B": "hello2",
	}
	type out struct {
		A int64
		B string
	}

	// Unpack struct slice
	var v []out
	err := mapper{m}.unpack(v)
	if err == nil {
		t.Fatal("should return error")
	}
	err = mapper{m}.unpack(&v)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = mapper{m2}.unpack(&v)
	if err != nil {
		t.Fatal(err.Error())
	}
	if x := len(v); x != 2 {
		t.Fatal(x)
	}
	if x := v[0].A; x != 1 {
		t.Fatal(x)
	}
	if x := v[1].A; x != 2 {
		t.Fatal(x)
	}
	if x := v[0].B; x != "hello" {
		t.Fatal(x)
	}
	if x := v[1].B; x != "hello2" {
		t.Fatal(x)
	}
}
