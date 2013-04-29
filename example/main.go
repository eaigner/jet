package main

import (
	jet ".."
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

func main() {
	// Open database
	db, err := jet.Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Reset
	err = db.Query(`DROP SCHEMA public CASCADE`).Run()
	if err != nil {
		panic(err)
	}
	err = db.Query(`CREATE SCHEMA public`).Run()
	if err != nil {
		panic(err)
	}

	// Set a logger
	logger := jet.NewLogger(os.Stdout)
	db.SetLogger(logger)

	// Create a migration suite
	var s jet.Suite
	s.AddSQL(
		`CREATE EXTENSION IF NOT EXISTS hstore`,
		`DROP EXTENSION IF EXISTS hstore`,
	)
	s.AddSQL(
		`CREATE TABLE "fruits" ( id serial, name text, attrs hstore )`,
		`DROP TABLE "fruits"`,
	)
	s.AddSQL(
		`CREATE INDEX "fruits_name_index" ON "fruits" ( "name" )`,
		`DROP INDEX "fruits_name_index"`,
	)

	// Run all migrations. The current migration version is stored, already applied migrations are not run twice!
	if err, _ := s.Migrate(db); err != nil {
		panic(err)
	}

	// Insert rows using a transaction
	txn, err := db.Begin()
	if err != nil {
		panic(err)
	}
	txn.Query(
		`INSERT INTO "fruits" ( "name", "attrs" ) VALUES ( $1, hstore(ARRAY[[$2, $3],[$4, $5]]) )`,
		"banana",
		"color", "yellow",
		"price", 2,
	).Run()
	txn.Query(
		`INSERT INTO "fruits" ( "name", "attrs" ) VALUES ( $1, hstore(ARRAY[[$2, $3],[$4, $5]]) )`,
		"orange",
		"color", "orange",
		"price", 1,
	).Run()
	txn.Query(
		`INSERT INTO "fruits" ( "name", "attrs" ) VALUES ( $1, hstore(ARRAY[[$2, $3],[$4, $5]]) )`,
		"grape",
		"color", "green",
	).Run()
	if err = txn.Commit(); err != nil {
		panic(err)
	}

	// Select some rows
	var fruits []struct {
		Name  string
		Color string
		Price string
	}
	if err := db.Query(
		`SELECT "name", attrs->'color' as "color", attrs->'price' as "price" FROM "fruits"`,
	).Rows(&fruits); err != nil {
		panic(err)
	}

	fmt.Println("FRUITS:", fruits)

	// Reset db
	if err, _ := s.Reset(db); err != nil {
		panic(err)
	}
}
