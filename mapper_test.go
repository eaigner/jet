package jet

import (
	"reflect"
	"testing"
)

func TestUnpackStruct(t *testing.T) {
	keys := []string{"ab_c", "c_d", "e", "f", "g"}
	vals := []interface{}{
		int64(9),
		"hello",
		"unsettable",
		[]uint8("uint8str"),
		[]uint8("uint8data"),
	}
	mppr := &mapper{
		keys:   keys,
		values: vals,
		conv:   SnakeCaseConverter,
	}

	// Unpack struct
	var v struct {
		AbC int64
		CD  string
		e   string
		F   string
		G   []byte
	}
	err := mppr.unpack(v)
	if err == nil {
		t.Fatal("should return error")
	}
	err = mppr.unpack(&v)
	if err != nil {
		t.Fatal(err)
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
	keys := []string{"ab_c", "c_d", "e"}
	vals := []interface{}{int64(9), "hello", "unsettable"}
	m := make(map[string]interface{})
	for i, k := range keys {
		m[k] = vals[i]
	}
	mppr := &mapper{
		keys:   keys,
		values: vals,
		conv:   SnakeCaseConverter,
	}

	var m2 map[string]interface{}
	err := mppr.unpack(&m2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, m2) {
		t.Log(m)
		t.Log(m2)
		t.Fatal("not equal")
	}
}

func TestUnpackStructSlice(t *testing.T) {
	k1 := []string{"A", "B"}
	v1 := []interface{}{int64(1), "hello"}
	v2 := []interface{}{int64(2), "hello2"}
	mppr := &mapper{
		keys:   k1,
		values: v1,
		conv:   SnakeCaseConverter,
	}

	// Unpack struct slice
	var v []struct {
		A int64
		B string
	}
	err := mppr.unpack(&v)
	if err != nil {
		t.Fatal(err.Error())
	}

	mppr.values = v2
	err = mppr.unpack(&v)
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
