package jet

import (
	"fmt"
	"reflect"
	"strings"
)

type mapper struct {
	columns map[string]interface{}
}

func (m mapper) unpack(v interface{}) error {
	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unpack result to Run(non-pointer %s)", vt.String())
	}
	return m.unpackValue(reflect.ValueOf(v))
}

func (m mapper) unpackValue(pv reflect.Value) error {
	switch pv.Kind() {
	case reflect.Ptr:
		return m.unpackValue(reflect.Indirect(pv))
	case reflect.Struct:
		return m.unpackStruct(pv)
	case reflect.Map:
		return m.unpackMap(pv)
	case reflect.Slice:
		sv := reflect.New(pv.Type().Elem())
		err := m.unpackValue(sv)
		if err != nil {
			return err
		}
		pv.Set(reflect.Append(pv, sv.Elem()))
		return nil
	}
	return fmt.Errorf("cannot unpack result to Run(%s %s)", pv.Type().String(), pv.Kind())
}

func (m mapper) unpackStruct(pv reflect.Value) error {
	v := reflect.Indirect(pv)
	for col, val := range m.columns {
		name := columnToFieldName(col)
		field := v.FieldByName(name)
		if field.IsValid() {
			field.Set(reflect.ValueOf(val))
		}
	}
	return nil
}

func (m mapper) unpackMap(pv reflect.Value) error {
	iv := reflect.Indirect(pv)
	mv := reflect.MakeMap(iv.Type())
	iv.Set(mv)
	for k, v := range m.columns {
		iv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return nil
}

func columnToFieldName(s string) string {
	name := ""
	if l := len(s); l > 0 {
		chunks := strings.Split(s, "_")
		for i, v := range chunks {
			chunks[i] = strings.Title(v)
		}
		name = strings.Join(chunks, "")
	}
	return name
}
