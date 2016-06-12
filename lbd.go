package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	//"os/signal"
	//"syscall"
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/reguero/golbd/lbcluster"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

type Config struct {
	Master          string
	HeartbeatFile   string
	HeartbeatPath   string
	HeartbeatMu     sync.Mutex
	TsigKeyPrefix   string
	TsigInternalKey string
	TsigExternalKey string
	SnmpPassword    string
	DnsManager      string
	Clusters        map[string][]string
	Parameters      map[string]lbcluster.Params
}

//func logInfo(log *syslog.Writer, s string) error {
//	//err := log.Info(s)
//	b := []byte(s)
//	_, err := log.Write(b)
//	fmt.Println(s)
//	return err
//}

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

func loadClusters(config Config) []lbcluster.LBCluster {
	var hm map[string]int
	var lbc lbcluster.LBCluster
	var lbcs []lbcluster.LBCluster

	for k, v := range config.Clusters {
		if len(v) == 0 {
			fmt.Println("cluster " + k + " is ignored as it has no members defined in the configuration file " + *configFileFlag)
			continue
		}
		if par, ok := config.Parameters[k]; ok {
			lbc = lbcluster.LBCluster{Cluster_name: k, Loadbalancing_username: "loadbalancing", Loadbalancing_password: config.SnmpPassword, Parameters: par, Statistics_filename: "/var/log/lb/lbstatistics." + k, Per_cluster_filename: "./" + k + ".log"}
			hm = make(map[string]int)
			for _, h := range v {
				hm[h] = lbcluster.WorstValue
			}
			lbc.Host_metric_table = hm
			lbcs = append(lbcs, lbc)
			fmt.Println("(re-)loaded cluster " + k)

		} else {
			fmt.Println("missing parameters for cluster " + k + "; ignoring the cluster, please check the configuration file " + *configFileFlag)
		}
	}
	return lbcs

}

func loadConfig(configFile string) (Config, error) {
	var config Config
	var p lbcluster.Params
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
			jsonStream := "{"
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
				fmt.Println(words[1])
				fmt.Println(words[3:])
			}
		}
	}
	config.Parameters = mp
	config.Clusters = mc
	return config, nil

}

func should_update_dns(config Config, hostname string, lg lbcluster.Log) bool {
	if hostname == config.Master {
		return true
	}
	master_heartbeat := "I am sick"
	connectTimeout := (10 * time.Second)
	readWriteTimeout := (20 * time.Second)
	httpClient := lbcluster.NewTimeoutClient(connectTimeout, readWriteTimeout)
	response, err := httpClient.Get("http://" + config.Master + "/load-balancing/" + config.HeartbeatFile)
	if err != nil {
		lg.Warning(fmt.Sprintf("problem fetching heartbeat file from the primary master %v: %v", config.Master, err))
		return true
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
		}
		fmt.Printf("%s", contents)
		master_heartbeat = strings.TrimSpace(string(contents))
		lg.Info("primary master heartbeat: " + master_heartbeat)
		r, _ := regexp.Compile(config.Master + ` : (\d+) : I am alive`)
		if r.MatchString(master_heartbeat) {
			matches := r.FindStringSubmatch(master_heartbeat)
			fmt.Println(matches[1])
			if mastersecs, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				now := time.Now()
				localsecs := now.Unix()
				diff := localsecs - mastersecs
				lg.Info(fmt.Sprintf("primary master heartbeat time difference: %v seconds", diff))
				if diff > 600 {
					return true
				}
			}
		} else {
			// Upload - heartbeat has unexpected values
			return true
		}
		// Do not upload, heartbeat was OK
		return false
	}
}

func update_heartbeat(config Config, hostname string, lg lbcluster.Log) error {
	if hostname != config.Master {
		return nil
	}
	heartbeat_file := config.HeartbeatPath + "/" + config.HeartbeatFile + "temp"
	heartbeat_file_real := config.HeartbeatPath + "/" + config.HeartbeatFile

	config.HeartbeatMu.Lock()
	defer config.HeartbeatMu.Unlock()

	f, err := os.OpenFile(heartbeat_file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		lg.Info(fmt.Sprintf("can not open %v for writing: %v", heartbeat_file, err))
		return err
	}
	now := time.Now()
	secs := now.Unix()
	_, err = fmt.Fprintf(f, "%v : %v : I am alive\n", hostname, secs)
	lg.Info("updating: heartbeat file " + heartbeat_file)
	if err != nil {
		lg.Info(fmt.Sprintf("can not write to %v: %v", heartbeat_file, err))
	}
	f.Close()
	if err = os.Rename(heartbeat_file, heartbeat_file_real); err != nil {
		lg.Info(fmt.Sprintf("can not rename %v to %v: %v", heartbeat_file, heartbeat_file_real, err))
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("This is a proof of concept golbd version %s \n", "0.000")
		os.Exit(0)
	}

	log, e := syslog.New(syslog.LOG_NOTICE, "lbd")
	lg := lbcluster.Log{*log, *debugFlag}
	if e == nil {
		lg.Info("Starting lbd")
	}

	hostname, e := os.Hostname()
	if e == nil {
		lg.Info("Hostname: " + hostname)
	}

	config, e := loadConfig(*configFileFlag)
	if e != nil {
		lg.Warning("loadConfig Error: ")
		lg.Warning(e.Error())
		os.Exit(1)
	} else {
		if *debugFlag {
			fmt.Println(config)
		}
	}

	if *debugFlag {
		for k, v := range config.Parameters {
			fmt.Println("params ", k, v)
		}
		for k, v := range config.Clusters {
			fmt.Println("clusters ", k, v)
		}
	}
	lbclusters := loadClusters(config)
	for _, c := range lbclusters {
		pc := &c
		pc.Slog = lg
		if *debugFlag {
			fmt.Println("lbcluster ", c)
		}
		if pc.Time_to_refresh() {
			pc.Find_best_hosts()
			if should_update_dns(config, hostname, lg) {
				fmt.Println("should_update_dns true")
				e = pc.Get_state_dns(config.DnsManager)
				if e != nil {
					lg.Warning("Get_state_dns Error: ")
					lg.Warning(e.Error())
				}
				e = pc.Update_dns(config.TsigKeyPrefix+"internal.", config.TsigInternalKey, config.DnsManager)
				if e != nil {
					lg.Warning("Internal Update_dns Error: ")
					lg.Warning(e.Error())
				}
				if pc.Externally_visible() {
					e = pc.Update_dns(config.TsigKeyPrefix+"external.", config.TsigExternalKey, config.DnsManager)
					if e != nil {
						lg.Warning("External Update_dns Error: ")
						lg.Warning(e.Error())
					}
				}
				update_heartbeat(config, hostname, lg)
			} else {
				fmt.Println("should_update_dns false")
			}
		}
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

	finish(done, &wg, lg)
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

func finish(done chan struct{}, wg *sync.WaitGroup, lg lbcluster.Log) {
	close(done)
	wg.Wait()
	lg.Info("all done!")
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
