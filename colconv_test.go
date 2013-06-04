package jet

import (
	"testing"
)

func TestSnakeColumnToFieldName(t *testing.T) {
	c := SnakeCaseConverter
	if x := c.ColumnToFieldName("column_one"); x != "ColumnOne" {
		t.Fatal(x)
	}
	if x := c.ColumnToFieldName("a"); x != "A" {
		t.Fatal(x)
	}
	if x := c.ColumnToFieldName("ab"); x != "Ab" {
		t.Fatal(x)
	}
	if x := c.ColumnToFieldName("Already_Camel"); x != "AlreadyCamel" {
		t.Fatal(x)
	}
}
