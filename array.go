package jet

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	markTestRx  = regexp.MustCompile(`(\$\d+)`)
	markSplitRx = regexp.MustCompile(`(\$\d+|\?)`)
)

func substituteMapAndArrayMarks(query string, args ...interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, 0, len(args)*2)
	newParts := []string{}

	markFormat := "?"
	markPlaceholder := "?"
	var usesNumberedMarkers = markTestRx.MatchString(query)
	if usesNumberedMarkers {
		markFormat = "$%d"
		markPlaceholder = "$0"
	}

	queryParts := markSplitRx.Split(query, -1)
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
	*newParts = append(*newParts, "hstore(ARRAY[ ")
	serializeSlice(markPlaceholder, reflect.ValueOf(a), newArgs, newParts)
	*newParts = append(*newParts, " ])")
}

func sanitizeMarkEnumeration(usesNumberedMarkers bool, markFormat, query string) string {
	parts := markSplitRx.Split(query, -1)
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
	for i, v := range a {
		uq, _ := strconv.Unquote(v)
		if i%2 == 0 {
			lastKey = uq
		} else {
			m[lastKey] = uq
		}
	}

	return m
}
