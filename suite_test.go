package jet

import (
	"reflect"
	"testing"
)

func TestBuildList(t *testing.T) {
	s := NewSuite()
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

func TestSuite(t *testing.T) {
	s := NewSuite()
	s.AddSQL(
		`CREATE TABLE "suite_test" ( id serial, name text )`,
		`DROP TABLE "suite_test"`,
	)
	s.AddSQL(
		`ALTER TABLE "suite_test" ADD "newcol" integer`,
		`ALTER TABLE "suite_test" DROP COLUMN "newcol"`,
	)
	s.AddSQL(
		`ALTER TABLE "suite_test" ADD "anothercol" bytea`,
		`ALTER TABLE "suite_test" DROP COLUMN "anothercol"`,
	)
	s.AddSQL(
		`CREATE INDEX "name_index" ON "suite_test" ( "name" )`,
		`DROP INDEX "name_index"`,
	)

	db, err := Open("postgres", "user=jet dbname=jet sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}
	db.SetLogger(NewLogger(nil))

	if err := db.Query(`DROP TABLE IF EXISTS "suite_test"`).Run(); err != nil {
		t.Fatal(err.Error())
	}
	if err := db.Query(`DROP TABLE IF EXISTS "migrations"`).Run(); err != nil {
		t.Fatal(err.Error())
	}

	if err := s.Run(db, true, 0); err != nil {
		t.Fatal(err)
	}

	if err := s.Run(db, false, 0); err != nil {
		t.Fatal(err)
	}
}
