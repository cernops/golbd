package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	//"os/signal"
	"sync"
	//"syscall"
	"time"
)

var versionFlag = flag.Bool("version", false, "print golbd version and exit")

func logInfo(log *syslog.Writer, s string) error {
	err := log.Info(s)
	fmt.Println(s)
	return err
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("This is a proof of concept golbd version %s \n", "0.000")
		os.Exit(0)
	}

	log, e := syslog.New(syslog.LOG_NOTICE, "lbd")
	if e == nil {
		logInfo(log, "Starting lbd")
	}

	var wg sync.WaitGroup
	done := make(chan struct{})
	wq := make(chan interface{})
	workerCount := 20
	//installSignalHandler(finish, done, &wg, log)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go doit(i, wq, done, &wg)
	}

	for i := 0; i < workerCount; i++ {
		wq <- i
	}

	finish(done, &wg, log)
}

func doit(workerId int, wq <-chan interface{}, done <-chan struct{}, wg *sync.WaitGroup) {
	fmt.Printf("[%v] is running\n", workerId)
	defer wg.Done()
	for {
		time.Sleep(3 * time.Second)
		select {
		case m := <-wq:
			fmt.Printf("[%v] m => %v\n", workerId, m)
		case <-done:
			fmt.Printf("[%v] is done\n", workerId)
			return
		}
	}
}

//type finishFunc func(chan struct{}, *sync.WaitGroup, *syslog.Writer)

func finish(done chan struct{}, wg *sync.WaitGroup, log *syslog.Writer) {
	close(done)
	wg.Wait()
	logInfo(log, "all done!")
	return
}

//func installSignalHandler(f finishFunc, done chan struct{}, wg *sync.WaitGroup, log *syslog.Writer) {
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
//
//	// Block until a signal is received.
//	go func() {
//		sig := <-c
//		mess := fmt.Sprintf("Exiting given signal: %v", sig)
//		logInfo(log, mess)
//		logInfo(log, "before exit")
//		f(done, wg, log)
//		logInfo(log, "about to exit")
//		os.Exit(0)
//	}()
//}
