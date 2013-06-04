package jet

import (
	"strings"
)

// ColumnConverter converts between struct field names
// and database row names
type ColumnConverter interface {
	ColumnToFieldName(col string) string
}

// SnakeCaseConverter converts column names from snake_case to CamelCase field names.
var SnakeCaseConverter ColumnConverter = &snakeConv{}

type snakeConv struct{}

func (conv *snakeConv) ColumnToFieldName(col string) string {
	name := ""
	if l := len(col); l > 0 {
		chunks := strings.Split(col, "_")
		for i, v := range chunks {
			chunks[i] = strings.Title(v)
		}
		name = strings.Join(chunks, "")
	}
	return name
}
