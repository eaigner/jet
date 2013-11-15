package jet

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	markRx = regexp.MustCompile(`(\$\d+|\?)`)
)

func substituteMapAndArrayMarks(query string, args ...interface{}) (string, []interface{}) {
	loc := markRx.FindStringIndex(query)
	if loc == nil {
		return query, args
	}
	usesNumberedMarkers := false
	markFormat := "?"
	markPlaceholder := "?"

	if query[loc[0]:loc[0]+1] == "$" {
		usesNumberedMarkers = true
		markFormat = "$%d"
		markPlaceholder = "$0"
	}

	newArgs := make([]interface{}, 0, len(args)*2)
	newParts := []string{}
	queryParts := markRx.Split(query, -1)
	for i, part := range queryParts {
		newParts = append(newParts, part)
		if i > len(args)-1 {
			break
		}
		arg := args[i]
		val := reflect.ValueOf(arg)
		k := val.Kind()

		if k == reflect.Map {
			serializeMap(markPlaceholder, val, &newArgs, &newParts)
		} else if k == reflect.Slice && val.Type() != reflect.TypeOf([]byte{}) {
			serializeSlice(markPlaceholder, val, &newArgs, &newParts)
		} else {
			newParts = append(newParts, markPlaceholder)
			newArgs = append(newArgs, arg)
		}
	}
	return sanitizeMarkEnumeration(usesNumberedMarkers, markFormat, strings.Join(newParts, "")), newArgs
}

func serializeSlice(markPlaceholder string, v reflect.Value, newArgs *[]interface{}, newParts *[]string) {
	a := make([]string, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		val := v.Index(i)
		a = append(a, markPlaceholder)
		*newArgs = append(*newArgs, val.Interface())
	}
	*newParts = append(*newParts, strings.Join(a, ", "))
}

func serializeMap(markPlaceholder string, v reflect.Value, newArgs *[]interface{}, newParts *[]string) {
	a := make([]interface{}, 0, v.Len()*2)
	for _, keyVal := range v.MapKeys() {
		val := v.MapIndex(keyVal)
		a = append(a, keyVal.Interface(), val.Interface())
	}
	serializeSlice(markPlaceholder, reflect.ValueOf(a), newArgs, newParts)
}

func sanitizeMarkEnumeration(usesNumberedMarkers bool, markFormat, query string) string {
	parts := markRx.Split(query, -1)
	a := make([]string, 0, len(parts)*2)
	for i, v := range parts {
		a = append(a, v)
		if i < len(parts)-1 {
			if usesNumberedMarkers {
				a = append(a, fmt.Sprintf(markFormat, i+1))
			} else {
				a = append(a, markFormat)
			}
		}
	}
	return strings.Join(a, "")
}

func parseHstoreColumn(s string) map[string]interface{} {
	lasti := 0
	quoteOpen := false
	escaped := false
	a := make([]string, 0, len(s))
	for i, r := range s {
		switch r {
		case '\\':
			escaped = true
		case 'N':
			if !quoteOpen && strings.HasPrefix(s[i:], "NULL") {
				a = append(a, "NULL")
			}
		case '"':
			if !escaped {
				quoteOpen = !quoteOpen
				if quoteOpen {
					lasti = i
				} else {
					a = append(a, s[lasti:i+1])
				}
			}
			escaped = false
		default:
			escaped = false
		}
	}

	if len(a)%2 == 1 {
		panic(fmt.Sprintf("invalid hstore map: %v", a))
	}

	// Convert to map
	m := make(map[string]interface{}, len(a)/2)
	lastKey := ""
	uq := ""
	isNull := false
	for i, v := range a {
		if v == "NULL" {
			isNull = true
		} else {
			uq, _ = strconv.Unquote(v)
			isNull = false
		}

		if i%2 == 0 {
			lastKey = uq
		} else if isNull {
			m[lastKey] = nil
		} else {
			m[lastKey] = uq
		}
	}

	return m
}
