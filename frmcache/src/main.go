package main

import (
	"flag"
	"fmt"
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
	flag.Parse()

	cacheWorker := NewCacheWorker("http://" + frmHostname + ":" + strconv.Itoa(frmPort))
	cacheWorker.Start()

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
