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
	"strings"
	"syscall"
)

func main() {
	var frmHostname string
	flag.StringVar(&frmHostname, "hostname", "localhost", "hostname of Ficsit Remote Monitoring webserver")
	var frmPort int
	flag.IntVar(&frmPort, "port", 8080, "port of Ficsit Remote Monitoring webserver")

	var frmHostnames string
	flag.StringVar(&frmHostnames, "hostnames", "", "comma separated values of multiple Ficsit Remote Monitoring webservers, of the form http://myserver1:8080,http://myserver2:8080. If defined, this will be used instead of hostname+port")

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

	cacheWorkers := []*CacheWorker{}
	if frmHostnames == "" {
		cacheWorkers = append(cacheWorkers, NewCacheWorker("http://"+frmHostname+":"+strconv.Itoa(frmPort), db))
	} else {
		for _, frmServer := range strings.Split(frmHostnames, ",") {
			if !strings.HasPrefix(frmServer, "http://") && !strings.HasPrefix(frmServer, "https://") {
				frmServer = "http://" + frmServer
			}
			cacheWorkers = append(cacheWorkers, NewCacheWorker(frmServer, db))
		}
	}
	for _, cacheWorker := range cacheWorkers {
		go cacheWorker.Start()
	}

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

	for _, cacheWorker := range cacheWorkers {
		cacheWorker.Stop()
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
