package main

import (
	jet ".."
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"runtime/pprof"
	"sync"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	run()
}

func run() {
	db, err := jet.Open("mysql", "benchmarkdbuser:benchmarkdbpass@tcp(localhost:3306)/hello_world?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`DROP TABLE IF EXISTS benchmark`).Run()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`CREATE TABLE benchmark ( a VARCHAR(40), b INT )`).Run()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`INSERT INTO benchmark ( a, b ) VALUES ( ?, ? )`, "benchme!", 9).Run()
	if err != nil {
		log.Fatal(err)
	}
	n := 20
	loops := 10000000
	nloops := loops / n

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < nloops; j++ {
				wg.Add(1)
				var v []struct {
					A string
					B int64
				}
				db.Query(`SELECT * FROM benchmark`).Rows(&v)
				wg.Done()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
