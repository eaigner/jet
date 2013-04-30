package jet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	markRx = regexp.MustCompile(`\$\d+`)
)

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
		case map[string]interface{}:
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
