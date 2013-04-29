package jet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	markRx     = regexp.MustCompile(`\$\d+`)
	hstorColRx = regexp.MustCompile(`(".*?[^\\]")=>(".*?[^\\]")`)
)

type Hstore map[string]interface{}

// hstore(ARRAY[$2, $3, $4, $5])

func substituteHstoreMarks(query string, args ...interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, 0, len(args)*2)
	newParts := []string{}
	mark := "$0"
	for i, part := range regexpSplit(markRx, query, -1) {
		newParts = append(newParts, part)
		if i > len(args)-1 {
			break
		}
		arg := args[i]
		switch t := arg.(type) {
		case Hstore:
			newParts = append(newParts, "hstore(ARRAY[")
			a := []string{}
			for k, v := range t {
				newArgs = append(newArgs, k, v)
				a = append(a, mark, mark)
			}
			newParts = append(newParts, strings.Join(a, ", "), "])")
		default:
			newParts = append(newParts, mark)
			newArgs = append(newArgs, arg)
		}
	}
	return sanitizeMarkEnumeration(strings.Join(newParts, "")), newArgs
}

func sanitizeMarkEnumeration(query string) string {
	parts := regexpSplit(markRx, query, -1)
	a := make([]string, 0, len(parts)*2)
	for i, v := range parts {
		a = append(a, v)
		if i < len(parts)-1 {
			a = append(a, fmt.Sprintf("$%d", i+1))
		}
	}
	return strings.Join(a, "")
}

func parseHstoreColumn(s string) Hstore {
	m := hstorColRx.FindAllStringSubmatch(s, -1)
	if len(m)%2 == 1 {
		panic("invalid hstore map")
	}
	h := make(Hstore)
	if len(m) > 0 {
		for _, v := range m {
			k, _ := strconv.Unquote(v[1])
			v, _ := strconv.Unquote(v[2])
			if len(k) > 0 {
				h[k] = v
			}
		}
	}
	return h
}

// TODO: replace when Go 1.1. is released
// This is the split source from http://tip.golang.org/src/pkg/regexp/regexp.go?s=33145:33194#L1067
func regexpSplit(re *regexp.Regexp, s string, n int) []string {
	if n == 0 {
		return nil
	}
	if len(s) == 0 {
		return []string{""}
	}

	matches := re.FindAllStringIndex(s, n)
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}
		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]
	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}
