package jet

import (
	_ "github.com/bmizerany/pq"
	"testing"
	"time"
)

func TestTx(t *testing.T) {
	db, err := Open("postgres", "user=jet dbname=jet sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = db.Query(`DROP TABLE IF EXISTS "tx_table"`).Run()
	if err != nil {
		t.Fatal(err.Error())
	}
	err = db.Query(`CREATE TABLE "tx_table" ( "a" text, "b" integer )`).Run()
	if err != nil {
		t.Fatal(err.Error())
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err.Error())
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
		t.Fatal(err.Error())
	}
	if x := len(tx.Errors()); x != 2 {
		t.Fatalf("should report %d errors, has %d", 2, x)
	}
}
