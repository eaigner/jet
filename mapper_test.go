package jet

import (
	"reflect"
	"testing"
)

func TestUnpackSimple(t *testing.T) {
	mppr := &mapper{
		conv: SnakeCaseConverter,
	}

	var out int
	err := mppr.unpack(nil, []interface{}{5}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out != 5 {
		t.Fatal(out)
	}

	// Double pointer
	var out2 *string
	err = mppr.unpack(nil, []interface{}{"hello!"}, &out2)
	if err != nil {
		t.Fatal(err)
	}
	if *out2 != "hello!" {
		t.Fatal(out)
	}
}

func TestUnpackNested(t *testing.T) {
	type A struct {
		B string
	}
	type C struct {
		A
		E int
	}

	keys := []string{"b", "e"}
	values := []interface{}{"b!", int(9)}
	mp := &mapper{
		conv: SnakeCaseConverter,
	}

	var out C
	err := mp.unpack(keys, values, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.B != "b!" {
		t.Fatal(out)
	}
	if out.E != 9 {
		t.Fatal(out)
	}
}

type custom struct {
	a string
	b string
}

func (c *custom) Encode() interface{} {
	return c.a + c.b
}

func (c *custom) Decode(v interface{}) error {
	if c == nil {
		c = new(custom)
	}
	s, ok := v.(string)
	if ok {
		c.a = string(s[0])
		c.b = string(s[1])
	}
	return nil
}

type plainCustom string

func (c *plainCustom) Encode() interface{} {
	return *c + "!"
}

func (c *plainCustom) Decode(v interface{}) error {
	s, ok := v.(string)
	if ok {
		*c = plainCustom(s)
	}
	return nil
}

func TestUnpackStruct(t *testing.T) {
	keys := []string{"ab_c", "c_d", "e", "f", "g", "h", "i", "j", "k", "l", "m"}
	sptr := "cd"
	vals := []interface{}{
		int64(9),
		"hello",
		"unsettable",
		[]uint8("uint8str"),
		[]uint8("uint8data"),
		[]uint8("1"),
		int64(1),
		"xy",
		"ab",
		&sptr,
		"s",
	}
	mppr := &mapper{
		conv: SnakeCaseConverter,
	}

	var v struct {
		AbC int64
		CD  string
		e   string
		F   string
		G   []byte
		H   bool
		I   bool
		J   *custom
		K   custom
		L   custom
		M   plainCustom
	}
	err := mppr.unpack(keys, vals, v)
	if err == nil {
		t.Fatal("should return error")
	}
	err = mppr.unpack(keys, vals, &v)
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
	if x := v.H; x != true {
		t.Fatal(x)
	}
	if x := v.I; x != true {
		t.Fatal(x)
	}
	if v.J == nil {
		t.Fatal()
	}
	if v.J.a != "x" {
		t.Fatal(v.J)
	}
	if v.J.b != "y" {
		t.Fatal(v.J)
	}
	if v.K.a != "a" {
		t.Fatal(v.K)
	}
	if v.K.b != "b" {
		t.Fatal(v.K)
	}
	if v.L.a != "c" {
		t.Fatal(v.L)
	}
	if v.L.b != "d" {
		t.Fatal(v.L)
	}
	if v.M != "s" {
		t.Fatal(v.M)
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
		conv: SnakeCaseConverter,
	}

	var out map[string]interface{}
	err := mppr.unpack(keys, vals, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, out) {
		t.Log(m)
		t.Log(out)
		t.Fatal("not equal")
	}

	// Pointer map
	a, b := 1, 2
	m2 := map[string]*int{
		"a": &a,
		"b": &b,
	}
	keys2 := []string{}
	vals2 := []interface{}{}
	for k, v := range m2 {
		keys2 = append(keys2, k)
		vals2 = append(vals2, v)
	}

	var out2 map[string]*int
	err = mppr.unpack(keys2, vals2, &out2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m2, out2) {
		t.Log(m2)
		t.Log(out2)
		t.Fail()
	}
}

func TestUnpackStructSlice(t *testing.T) {
	k1 := []string{"A", "B"}
	v1 := []interface{}{int64(1), "hello"}
	v2 := []interface{}{int64(2), "hello2"}
	mppr := &mapper{
		conv: SnakeCaseConverter,
	}

	type styp struct {
		A int64
		B string
	}

	// Unpack struct slice
	var out []styp
	err := mppr.unpack(k1, v1, &out)
	if err != nil {
		t.Fatal(err)
	}

	err = mppr.unpack(k1, v2, &out)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(out); x != 2 {
		t.Fatal(x)
	}
	if x := out[0].A; x != 1 {
		t.Fatal(x)
	}
	if x := out[1].A; x != 2 {
		t.Fatal(x)
	}
	if x := out[0].B; x != "hello" {
		t.Fatal(x)
	}
	if x := out[1].B; x != "hello2" {
		t.Fatal(x)
	}

	// Unpack pointer slice
	var out2 []*styp
	err = mppr.unpack(k1, v1, &out2)
	if err != nil {
		t.Fatal(err)
	}
	if x := len(out2); x != 1 {
		t.Fatal(x)
	}
	if x := out2[0].A; x != 1 {
		t.Fatal(x)
	}
	if x := out2[0].B; x != "hello" {
		t.Fatal(x)
	}
}
