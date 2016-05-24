package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	//"os/signal"
	"sync"
	//"syscall"
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/reguero/golbd/lbcluster"
	"io"
	"strconv"
	"strings"
	"time"
)

var versionFlag = flag.Bool("version", false, "print golbd version and exit")
var debugFlag = flag.Bool("debug", false, "set golbd in debug mode")
var startFlag = flag.Bool("start", false, "start golbd")
var stopFlag = flag.Bool("stop", false, "stop golbd")
var updateFlag = flag.Bool("update", false, "update golbd config")
var configFileFlag = flag.String("config", "./load-balancing.conf", "specify configuration file path")
var logFileFlag = flag.String("log", "./lbd.log", "specify log file path")

const itCSgroupDNSserver string = "cfmgr.cern.ch"
const worstValue int = 99999

type Config struct {
	Master          string
	HeartbeatFile   string
	HeartbeatPath   string
	TsigKeyPrefix   string
	TsigInternalKey string
	TsigExternalKey string
	SnmpPassword    string
	DnsManager      string
	Clusters        map[string][]string
	Parameters      map[string]lbcluster.Params
}

func logInfo(log *syslog.Writer, s string) error {
	//err := log.Info(s)
	b := []byte(s)
	_, err := log.Write(b)
	fmt.Println(s)
	return err
}

// Read a whole file into the memory and store it as array of lines
func readLines(path string) (lines []string, err error) {
	var (
		file   *os.File
		part   []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	var p lbcluster.Params
	var jsonStream string = "{"
	var mc map[string][]string
	mc = make(map[string][]string)
	var mp map[string]lbcluster.Params
	mp = make(map[string]lbcluster.Params)

	lines, err := readLines(configFile)
	if err != nil {
		return config, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || (line == "") {
			continue
		}
		words := strings.Split(line, " ")
		if words[1] == "=" {
			switch words[0] {
			case "master":
				config.Master = words[2]
			case "heartbeat_path":
				config.HeartbeatPath = words[2]
			case "heartbeat_file":
				config.HeartbeatFile = words[2]
			case "tsig_key_prefix":
				config.TsigKeyPrefix = words[2]
			case "tsig_internal_key":
				config.TsigInternalKey = words[2]
			case "tsig_external_key":
				config.TsigExternalKey = words[2]
			case "snmpd_password":
				config.SnmpPassword = words[2]
			case "dns_manager":
				config.DnsManager = words[2]
			}
		} else if words[2] == "=" {
			if words[0] == "parameters" {
				for i, param := range words[3:] {
					keyval := strings.Split(param, "#")
					if keyval[1] == "no" {
						jsonStream = jsonStream + strconv.Quote(strings.Title(keyval[0])) + ": false"
					} else if keyval[1] == "yes" {
						jsonStream = jsonStream + strconv.Quote(strings.Title(keyval[0])) + ": true"
					} else if _, err := strconv.Atoi(keyval[1]); err == nil {
						jsonStream = jsonStream + strconv.Quote(strings.Title(keyval[0])) + ": " + keyval[1]
					} else {
						jsonStream = jsonStream + strconv.Quote(strings.Title(keyval[0])) + ": " + strconv.Quote(keyval[1])
					}
					if i < (len(words[3:]) - 1) {
						jsonStream = jsonStream + ", "
					}
				}
				jsonStream = jsonStream + "}"
				dec := json.NewDecoder(strings.NewReader(jsonStream))
				if err := dec.Decode(&p); err == io.EOF {
					break
				} else if err != nil {
					//log.Fatal(err)
					fmt.Println(err)
					os.Exit(1)
				}
				mp[words[1]] = p

			} else if words[0] == "clusters" {
				mc[words[1]] = words[3:]
			}
		}
	}
	config.Parameters = mp
	config.Clusters = mc
	return config, nil

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

	hostname, e := os.Hostname()
	if e == nil {
		logInfo(log, "Hostname: "+hostname)
	}

	config, e := loadConfig(*configFileFlag)
	if e != nil {
		fmt.Println("loadConfig Error: ", e)
		os.Exit(1)
	} else {
		fmt.Println(config)
	}

	for k, v := range config.Parameters {
		fmt.Println("params ", k, v)
	}
	for k, v := range config.Clusters {
		fmt.Println("clusters ", k, v)
	}
	os.Exit(0)
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
