package jet

import (
	"reflect"
	"testing"
)

func TestBuildList(t *testing.T) {
	var s Suite
	s.AddSQL("U1", "D1")
	s.AddSQL("U2", "D2")
	s.AddSQL("U3", "D3")
	s.AddSQL("U4", "D4")

	// Test up
	if l := s.buildList(true, 0); !reflect.DeepEqual(l, s.Migrations) {
		t.Fatal(l)
	}
	if l := s.buildList(true, 1); !reflect.DeepEqual(l, s.Migrations[1:]) {
		t.Fatal(l)
	}
	if l := s.buildList(true, 4); !reflect.DeepEqual(l, s.Migrations[4:]) {
		t.Fatal(l)
	}
	if l := s.buildList(true, 5); !reflect.DeepEqual(l, s.Migrations[4:]) {
		t.Fatal(l)
	}

	// Test down
	if l := s.buildList(false, 0); !reflect.DeepEqual(l, []*Migration{}) {
		t.Fatal(l)
	}
	if l := s.buildList(false, 1); !reflect.DeepEqual(l, reverse(s.Migrations[:1])) {
		t.Fatal(l)
	}
	if l := s.buildList(false, 4); !reflect.DeepEqual(l, reverse(s.Migrations[:4])) {
		t.Fatal(l)
	}
	if l := s.buildList(false, 5); !reflect.DeepEqual(l, reverse(s.Migrations[:4])) {
		t.Fatal(l)
	}
}
