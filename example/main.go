package main

import (
	jet ".."
	"fmt"
	_ "github.com/bmizerany/pq"
	"os"
)

func main() {
	// Open database
	db, err := jet.Open("postgres", "user=jet dbname=jet sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Set a logger
	logger := jet.NewLogger(os.Stdout)
	db.SetLogger(logger)

	// Create a migration suite
	var s jet.Suite
	s.AddSQL(
		`CREATE TABLE "fruits" ( id serial, name text )`,
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
	txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "banana").Run()
	txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "orange").Run()
	txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "grape").Run()
	if err = txn.Commit(); err != nil {
		panic(err)
	}

	// Select some rows
	var fruits []struct {
		Name string
	}
	if err := db.Query(`SELECT * FROM "fruits"`).Rows(&fruits); err != nil {
		panic(err)
	}

	fmt.Println("FRUITS:", fruits)

	// Reset db
	if err, _ := s.Reset(db); err != nil {
		panic(err)
	}
}
