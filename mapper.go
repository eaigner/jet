package jet

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type mapper struct {
	keys   []string
	values []interface{}
	conv   ColumnConverter
}

func (m *mapper) unpack(v interface{}) error {
	pv := reflect.ValueOf(v)
	if pv.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unpack result to non-pointer (%s)", pv.Type().String())
	}
	return m.unpackValue(pv)
}

func (m *mapper) unpackValue(pv reflect.Value) error {
	switch pv.Kind() {
	case reflect.Ptr:
		return m.unpackValue(pv.Elem())
	case reflect.Struct:
		return m.unpackStruct(pv)
	case reflect.Map:
		return m.unpackMap(pv)
	case reflect.Slice:
		sv := reflect.New(pv.Type().Elem()).Elem()
		err := m.unpackValue(sv)
		if err != nil {
			return err
		}
		pv.Set(reflect.Append(pv, sv))
		return nil
	}
	return fmt.Errorf("cannot unpack result to %s (%s)", pv.Type().String(), pv.Kind())
}

func (m *mapper) unpackStruct(pv reflect.Value) error {
	iv := reflect.Indirect(pv)
	for i, k := range m.keys {
		v := m.values[i]
		var name string
		if m.conv == nil {
			name = strings.ToUpper(k[:1]) + k[1:]
		} else if m.conv != nil {
			name = m.conv.ColumnToFieldName(k)
		}
		if f := iv.FieldByName(name); f.IsValid() {
			setValue(reflect.Indirect(reflect.ValueOf(v)), f)
		}
	}
	return nil
}

func (m *mapper) unpackMap(pv reflect.Value) error {
	iv := reflect.Indirect(pv)
	mv := reflect.MakeMap(iv.Type())
	iv.Set(mv)
	for i, k := range m.keys {
		v := m.values[i]
		iv.SetMapIndex(reflect.ValueOf(k), reflect.Indirect(reflect.ValueOf(v)))
	}
	return nil
}

func convertAndSet(f interface{}, to reflect.Value) {
	to.Set(reflect.ValueOf(f).Convert(to.Type()))
}

func setValue(from, to reflect.Value) {
	switch t := from.Interface().(type) {
	case []uint8:
		switch to.Interface().(type) {
		case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
			n, _ := strconv.ParseInt(string(t), 10, 64)
			convertAndSet(n, to)
		case string:
			to.SetString(string(t))
		case map[string]interface{}:
			to.Set(reflect.ValueOf(parseHstoreColumn(string(t))))
		default:
			convertAndSet(t, to)
		}
	default:
		convertAndSet(t, to)
	}
}
