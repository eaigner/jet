package jet

import (
	_ "github.com/lib/pq"
	"os"
	"testing"
	"time"
)

func TestTx(t *testing.T) {
	db, err := Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	l := NewLogger(os.Stdout)
	db.SetLogger(l)
	if db.Logger() != l {
		t.Fatal("wrong logger set")
	}
	err = db.Query(`DROP TABLE IF EXISTS "tx_table"`).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "tx_table" ( "a" text, "b" integer )`).Run()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if tx.Logger() != db.Logger() {
		t.Fatal("wrong logger set")
	}
	err1 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello", 7).Run()
	if err1 != nil {
		t.Fatal(err1.Error())
	}
	err2 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello2", time.Now()).Run()
	if err2 == nil {
		t.Fatal("should return error")
	}
	err3 := tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "hello2", "boo").Run()
	if err3 == nil {
		t.Fatal("should return error")
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	// Rollback
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Query(`INSERT INTO "tx_table" ( "a", "b" ) VALUES ( $1, $2 )`, "roll-me-back", 14).Run()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}
	var c int64
	err = db.Query(`SELECT COUNT(*) FROM "tx_table"`).Rows(&c)
	if err != nil {
		t.Fatal(err)
	}
	if c != 0 {
		t.Fatal(c)
	}
}
