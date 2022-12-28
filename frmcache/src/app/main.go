package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

func main() {
	var frmHostname string
	flag.StringVar(&frmHostname, "hostname", "localhost", "hostname of Ficsit Remote Monitoring webserver")
	var frmPort int
	flag.IntVar(&frmPort, "port", 8080, "port of Ficsit Remote Monitoring webserver")

	var pgHost string
	flag.StringVar(&pgHost, "pghost", "postgres", "postgres hostname")
	var pgPort int
	flag.IntVar(&pgPort, "pgport", 5432, "postgres port")
	var pgPassword string
	flag.StringVar(&pgPassword, "pgpassword", "secretpassword", "postgres password")
	var pgUser string
	flag.StringVar(&pgUser, "pguser", "postgres", "postgres username")
	var pgDb string
	flag.StringVar(&pgDb, "pgdb", "postgres", "postgres db")
	flag.Parse()

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPassword, pgDb)
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	defer db.Close()
	err = db.Ping()
	CheckError(err)

	cacheWorker := NewCacheWorker("http://"+frmHostname+":"+strconv.Itoa(frmPort), db)
	go cacheWorker.Start()

	fmt.Printf(`
FRM Cache started
Press ctrl-c to exit`)

	// Wait for an interrupt signal
	sigChan := make(chan os.Signal, 1)
	if runtime.GOOS == "windows" {
		signal.Notify(sigChan, os.Interrupt)
	} else {
		signal.Notify(sigChan, syscall.SIGTERM)
		signal.Notify(sigChan, syscall.SIGINT)
	}
	<-sigChan

	cacheWorker.Stop()
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
