package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	//"os/signal"
	//"syscall"
)

var versionFlag = flag.Bool("version", false, "print golbd version and exit")

func main() {
	flag.Parse()

	if versionFlag {
		fmt.Printf("This is a proof of concept golbd version %s \n", "0.000")
		os.Exit(0)
	}

	log, e := syslog.New(syslog.LOG_NOTICE, "lbd")
	if e == nil {
		log.Info("Starting lbd")
	}
}
