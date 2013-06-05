package main

import (
	jet ".."
	"flag"
	_ "github.com/lib/pq"
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
	log.Println("A")
	db, err := jet.Open("postgres", "user=postgres dbname=jet sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`DROP TABLE IF EXISTS "benchmark"`).Run()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`CREATE TABLE "benchmark" ( "a" text, "b" integer )`).Run()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Query(`INSERT INTO "benchmark" ( "a", "b" ) VALUES ( $1, $2 )`, "benchme!", 9).Run()
	if err != nil {
		log.Fatal(err)
	}
	n := 20
	loops := 1000000
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
				err = db.Query(`SELECT * FROM "table"`).Rows(&v)
				if err != nil {
					log.Fatal(err)
				}
				wg.Done()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("B")
}
