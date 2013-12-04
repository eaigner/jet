package jet

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type mapper struct {
	conv ColumnConverter
}

func (m *mapper) unpack(keys []string, values []interface{}, out interface{}) error {
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unpack result to non-pointer (%s)", val.Type().String())
	}
	return m.unpackValue(keys, values, val)
}

func (m *mapper) unpackValue(keys []string, values []interface{}, out reflect.Value) error {
	switch out.Interface().(type) {
	case ComplexValue:
		if out.IsNil() {
			out.Set(reflect.New(out.Type().Elem()))
		}
		plain := reflect.Indirect(reflect.ValueOf(values[0]))
		return out.Interface().(ComplexValue).Decode(plain.Interface())
	}
	if out.CanAddr() {
		switch out.Addr().Interface().(type) {
		case ComplexValue:
			return m.unpackValue(keys, values, out.Addr())
		}
	}
	switch out.Kind() {
	case reflect.Ptr:
		if out.IsNil() {
			out.Set(reflect.New(out.Type().Elem()))
		}
		return m.unpackValue(keys, values, reflect.Indirect(out))
	case reflect.Slice:
		if keys == nil {
			return m.unpackSimple(nil, values, out)
		} else {
			return m.unpackSlice(keys, values, out)
		}
	case reflect.Struct:
		return m.unpackStruct(keys, values, out)
	case reflect.Map:
		if keys == nil {
			return m.unpackSimple(nil, values, out)
		} else {
			return m.unpackMap(keys, values, out)
		}
	default:
		return m.unpackSimple(nil, values, out)
	}
	return fmt.Errorf("cannot unpack result to %T (%s)", out, out.Kind())
}

func (m *mapper) unpackSlice(keys []string, values []interface{}, out reflect.Value) error {
	elemTyp := reflect.Indirect(reflect.New(out.Type().Elem()))
	m.unpackValue(keys, values, elemTyp)
	out.Set(reflect.Append(out, elemTyp))
	return nil
}

func (m *mapper) unpackStruct(keys []string, values []interface{}, out reflect.Value) error {
	// If no keys are passed in it's probably a struct field requiring
	// a simple unpack like a time.Time struct field.
	if len(keys) == 0 {
		return m.unpackSimple(nil, values, out)
	}
	for i, k := range keys {
		var convKey string
		if m.conv == nil {
			convKey = strings.ToUpper(k[:1]) + k[1:]
		} else if m.conv != nil {
			convKey = m.conv.ColumnToFieldName(k)
		}
		field := out.FieldByName(convKey)
		if field.IsValid() {
			m.unpackValue(nil, values[i:i+1], field)
		}
	}
	return nil
}

func (m *mapper) unpackMap(keys []string, values []interface{}, out reflect.Value) error {
	if out.IsNil() {
		out.Set(reflect.MakeMap(out.Type()))
	}
	for i, k := range keys {
		elemTyp := reflect.Indirect(reflect.New(out.Type().Elem()))
		m.unpackValue(nil, values[i:i+1], elemTyp)
		out.SetMapIndex(reflect.ValueOf(k), elemTyp)
	}

	return nil
}

func (m *mapper) unpackSimple(keys []string, values []interface{}, out reflect.Value) error {
	if !out.IsValid() {
		panic("cannot unpack to zero value")
	}
	if len(values) != 1 {
		panic("cannot unpack to simple value, invalid values input")
	}
	setValue(reflect.Indirect(reflect.ValueOf(values[0])), out)
	return nil
}

func convertAndSet(f interface{}, to reflect.Value) {
	from := reflect.ValueOf(f)
	if from.IsValid() {
		to.Set(from.Convert(to.Type()))
	} else {
		to.Set(reflect.Zero(to.Type()))
	}
}

func setValue(from, to reflect.Value) {
	switch t := from.Interface().(type) {
	case []uint8:
		setValueFromBytes(t, to)
	case int, int8, int16, int32, int64:
		setValueFromInt(reflect.ValueOf(t).Int(), to)
	case uint, uint8, uint16, uint32, uint64:
		setValueFromUint(reflect.ValueOf(t).Uint(), to)
	case float32, float64:
		setValueFromFloat(reflect.ValueOf(t).Float(), to)
	case time.Time:
		setValueFromTime(t, to)
	default:
		convertAndSet(t, to)
	}
}

func setValueFromBytes(t []uint8, to reflect.Value) {
	switch to.Interface().(type) {
	case bool:
		n, _ := strconv.ParseInt(string(t), 10, 32)
		convertAndSet(bool(n == 1), to)
	case int, int8, int16, int32, int64:
		n, _ := strconv.ParseInt(string(t), 10, 64)
		convertAndSet(n, to)
	case uint, uint8, uint16, uint32, uint64:
		n, _ := strconv.ParseUint(string(t), 10, 64)
		convertAndSet(n, to)
	case float32:
		n, _ := strconv.ParseFloat(string(t), 32)
		convertAndSet(n, to)
	case float64:
		n, _ := strconv.ParseFloat(string(t), 64)
		convertAndSet(n, to)
	case string:
		to.SetString(string(t))
	case map[string]interface{}:
		to.Set(reflect.ValueOf(parseHstoreColumn(string(t))))
	default:
		convertAndSet(t, to)
	}
}

func setValueFromFloat(f float64, to reflect.Value) {
	switch to.Interface().(type) {
	case bool:
		convertAndSet(bool(f == 1), to)
	default:
		convertAndSet(f, to)
	}
}

func setValueFromInt(i int64, to reflect.Value) {
	switch to.Interface().(type) {
	case bool:
		convertAndSet(bool(i == 1), to)
	default:
		convertAndSet(i, to)
	}
}

func setValueFromUint(i uint64, to reflect.Value) {
	switch to.Interface().(type) {
	case bool:
		convertAndSet(bool(i == 1), to)
	default:
		convertAndSet(i, to)
	}
}

func setValueFromTime(t time.Time, to reflect.Value) {
	switch to.Interface().(type) {
	case int64:
		convertAndSet(int64(t.Unix()), to)
	case uint64:
		convertAndSet(uint64(t.Unix()), to)
	default:
		convertAndSet(t, to)
	}
}
