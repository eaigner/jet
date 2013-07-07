package jet

import (
	"testing"
)

func createDb(t *testing.T) Db {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db.SetLogger(NewLogger(nil))

	err = db.Query("DROP SCHEMA public CASCADE").Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query("CREATE SCHEMA public").Run()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestSuite(t *testing.T) {
	var s Suite
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

	db := createDb(t)

	if err, c, s := s.Run(db, true, 0); err != nil || c != 4 || s != 4 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Run(db, false, 0); err != nil || c != 0 || s != 4 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Migrate(db); err != nil || c != 4 || s != 4 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Reset(db); err != nil || c != 0 || s != 4 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Step(db); err != nil || c != 1 || s != 1 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Step(db); err != nil || c != 2 || s != 1 {
		t.Fatal(err, c, s)
	}
	if err, c, s := s.Rollback(db); err != nil || c != 1 || s != 1 {
		t.Fatal(err, c, s)
	}
}

func TestSuiteError(t *testing.T) {
	var s Suite
	s.AddSQL(`CREATE TABLE flowers (id serial)`, `DROP TABLE flowers`)
	s.AddSQL(`CREATE TABLE err ors (id serial)`, `DROP TABLE errors`)

	db := createDb(t)

	err, cur, applied := s.Migrate(db)
	if err == nil {
		t.Fatal("expected error")
	} else {
		t.Log(err)
	}
	if cur != 1 {
		t.Fatal(cur)
	}
	if applied != 1 {
		t.Fatal(applied)
	}
}
