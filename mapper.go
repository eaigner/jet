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
		return fmt.Errorf("cannot unpack result to non-pointer (%s)", vt.String())
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
	return fmt.Errorf("cannot unpack result to %s (%s)", pv.Type().String(), pv.Kind())
}

func (m mapper) unpackStruct(pv reflect.Value) error {
	iv := reflect.Indirect(pv)
	for k, v := range m.columns {
		name := columnToFieldName(k)
		field := iv.FieldByName(name)
		if field.IsValid() {
			target := reflect.Indirect(reflect.ValueOf(v)).Interface()
			switch t := target.(type) {
			case []uint8:
				if field.Kind() == reflect.String {
					field.SetString(string(t))
				} else {
					field.Set(reflect.ValueOf(target))
				}
			default:
				field.Set(reflect.ValueOf(target))
			}
		}
	}
	return nil
}

func (m mapper) unpackMap(pv reflect.Value) error {
	iv := reflect.Indirect(pv)
	mv := reflect.MakeMap(iv.Type())
	iv.Set(mv)
	for k, v := range m.columns {
		iv.SetMapIndex(reflect.ValueOf(k), reflect.Indirect(reflect.ValueOf(v)))
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
