Jet is a super-flexible lightweight SQL interface

## Example

```go
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
err, _ = s.Migrate(db)
if err != nil {
  panic(err)
}

// Insert a row using a transaction
txn, err := db.Begin()
if err != nil {
  panic(err)
}
txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "banana").Run()
txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "orange").Run()
txn.Query(`INSERT INTO "fruits" ( "name" ) VALUES ( $1 )`, "grape").Run()
err = txn.Commit()
if err != nil {
  panic(err)
}

// Select some rows
var fruits []struct {
  Name string
}
err = db.Query(`SELECT * FROM "fruits"`).Rows(&fruits)
if err != nil {
  panic(err)
}

fmt.Println("FRUITS:", fruits)

// Reset db
err, _ = s.Reset(db)
if err != nil {
  panic(err)
}
```
