package jet

import (
	"fmt"
	"reflect"
	"strings"
)

type mapper struct {
	columns map[string]interface{}
	conv    ColumnConverter
}

func (m mapper) unpack(v interface{}) error {
	pv := reflect.ValueOf(v)
	if pv.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unpack result to non-pointer (%s)", pv.Type().String())
	}
	return m.unpackValue(pv)
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
		var name string
		if m.conv == nil {
			name = strings.ToUpper(k[:1]) + k[1:]
		} else if m.conv != nil {
			name = m.conv.ColumnToFieldName(k)
		}
		field := iv.FieldByName(name)
		if field.IsValid() {
			setValue(reflect.Indirect(reflect.ValueOf(v)).Interface(), field)
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

func setValue(i interface{}, v reflect.Value) {
	switch t := i.(type) {
	case []uint8:
		switch v.Interface().(type) {
		case string:
			v.SetString(string(t))
		case map[string]interface{}:
			v.Set(reflect.ValueOf(parseHstoreColumn(string(t))))
		default:
			v.Set(reflect.ValueOf(i))
		}
	default:
		v.Set(reflect.ValueOf(i).Convert(v.Type()))
	}
}
