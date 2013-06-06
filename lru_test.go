package jet

import (
	"database/sql"
	"testing"
)

func TestLRU(t *testing.T) {
	s1 := new(sql.Stmt)
	s2 := new(sql.Stmt)
	s3 := new(sql.Stmt)
	s4 := new(sql.Stmt)

	lru := newLRUCache(3)
	lru.set("one", s1)
	lru.set("two", s2)
	lru.set("two", s2)
	lru.set("three", s3)
	lru.set("four", s4)

	if x := len(lru.m); x != 3 {
		t.Fatal(x)
	}
	if x := lru.get("one"); x != nil {
		t.Fatal(lru.m)
	}
	if x := lru.get("two"); x != s2 {
		t.Fatal(lru.m)
	}
	if x := lru.get("three"); x != s3 {
		t.Fatal(lru.m)
	}
	if x := lru.get("four"); x != s4 {
		t.Fatal(lru.m)
	}
}
