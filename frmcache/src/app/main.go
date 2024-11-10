package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	_ "github.com/lib/pq"
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
	frmPort, _ := os.LookupEnv("FRM_PORT")
	frmHostnames, _ := os.LookupEnv("FRM_HOSTS")

	pgHost := lookupEnvWithDefault("PG_HOST", "postgres")
	pgPort, err := strconv.Atoi(lookupEnvWithDefault("PG_PORT", "5432"))
	if err != nil {
		pgPort = 5432
	}
	pgPassword := lookupEnvWithDefault("PG_PASSWORD", "secretpassword")
	pgUser := lookupEnvWithDefault("PG_USER", "postgres")
	pgDb := lookupEnvWithDefault("PG_DB", "postgres")
	migrationLocation := lookupEnvWithDefault("MIGRATION_DIR", "/var/lib/frmcache")

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPassword, pgDb)
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	defer db.Close()
	retries := 5
	for i := 0; i < retries; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("connecting to database...")
		time.Sleep(2 * time.Second)
	}
	CheckError(err)

	var m *migrate.Migrator
	migrateConn, err := pgx.Connect(context.Background(), psqlconn)
	if err != nil {
		log.Printf("Unable to establish connection: %v", err)
		return
	}
	m, err = migrate.NewMigrator(context.Background(), migrateConn, "schema_version")
	if err != nil {
		log.Printf("Unable to create migrator: %v", err)
		return
	}
	m.LoadMigrations(os.DirFS(migrationLocation))
	m.OnStart = func(_ int32, name, direction, _ string) {
		log.Printf("Migrating %s: %s", direction, name)
	}
	if err = m.Migrate(context.Background()); err != nil {
		log.Printf("Unexpected failure migrating: %v", err)
		return
	}

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

	log.Printf(`
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
