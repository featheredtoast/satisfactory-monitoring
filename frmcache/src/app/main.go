package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func lookupEnvWithDefault(variable string, defaultVal string) string {
	val, exist := os.LookupEnv(variable)
	if exist {
		return val
	}
	return defaultVal
}

func main() {
	frmHostname, _ := os.LookupEnv("FRM_HOST")
	frmPort, _ := os.LookupEnv("FRM_HOST")
	frmHostnames, _ := os.LookupEnv("FRM_HOSTS")

	pgHost := lookupEnvWithDefault("PG_HOST", "postgres")
	pgPort, err := strconv.Atoi(lookupEnvWithDefault("PG_HOST", "5432"))
	if err != nil {
		pgPort = 5432
	}
	pgPassword := lookupEnvWithDefault("PG_PASSWORD", "secretpassword")
	pgUser := lookupEnvWithDefault("PG_USER", "postgres")
	pgDb := lookupEnvWithDefault("PG_DB", "postgres")

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPassword, pgDb)
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	defer db.Close()
	err = db.Ping()
	CheckError(err)

	cacheWorkers := []*CacheWorker{}
	if frmHostnames == "" {
		cacheWorkers = append(cacheWorkers, NewCacheWorker("http://"+frmHostname+":"+frmPort, db))
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
