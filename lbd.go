package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

var versionFlag = flag.Bool("version", false, "print lbd version and exit")
var debugFlag = flag.Bool("debug", false, "set lbd in debug mode")
var startFlag = flag.Bool("start", false, "start lbd")
var stopFlag = flag.Bool("stop", false, "stop lbd")
var updateFlag = flag.Bool("update", false, "update lbd config")
var configFileFlag = flag.String("config", "./load-balancing.conf", "specify configuration file path")
var logFileFlag = flag.String("log", "./lbd.log", "specify log file path")
var stdoutFlag = flag.Bool("stdout", false, "send log to stdtout")

const itCSgroupDNSserver string = "cfmgr.cern.ch"

// Config this is the configuration of the lbd
type Config struct {
	Master          string
	HeartbeatFile   string
	HeartbeatPath   string
	HeartbeatMu     sync.Mutex
	TsigKeyPrefix   string
	TsigInternalKey string
	TsigExternalKey string
	SnmpPassword    string
	DNSManager      string
	Clusters        map[string][]string
	Parameters      map[string]lbcluster.Params
}

// readLines reads a whole file into memory and returns a slice of lines.
func readLines(path string) (lines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func loadClusters(config *Config, lg *lbcluster.Log) ([]lbcluster.LBCluster, error) {
	var lbc lbcluster.LBCluster
	var lbcs []lbcluster.LBCluster

	for k, v := range config.Clusters {
		if len(v) == 0 {
			lg.Warning("cluster: " + k + " ignored as it has no members defined in the configuration file " + *configFileFlag)
			continue
		}
		if par, ok := config.Parameters[k]; ok {
			lbc = lbcluster.LBCluster{Cluster_name: k, Loadbalancing_username: "loadbalancing",
				Loadbalancing_password: config.SnmpPassword, Parameters: par,
				Current_best_hosts:      []string{"unknown"},
				Previous_best_hosts:     []string{"unknown"},
				Previous_best_hosts_dns: []string{"unknown"},
				Slog:                    lg}
			hm := make(map[string]int)
			for _, h := range v {
				hm[h] = 100000
			}
			lbc.Host_metric_table = hm
			lbcs = append(lbcs, lbc)
			lbc.Write_to_log("INFO", "(re-)loaded cluster ")

		} else {
			lg.Warning("cluster: " + k + " missing parameters for cluster; ignoring the cluster, please check the configuration file " + *configFileFlag)
		}
	}

	return lbcs, nil

}

func loadConfig(configFile string, lg *lbcluster.Log) (*Config, []lbcluster.LBCluster, error) {
	var (
		config Config
		p      lbcluster.Params
		mc     = make(map[string][]string)
		mp     = make(map[string]lbcluster.Params)
	)

	lines, err := readLines(configFile)
	if err != nil {
		return nil, nil, err
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
				config.DNSManager = words[2]
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
					lg.Warning(fmt.Sprintf("%v", err))
					os.Exit(1)
				}
				mp[words[1]] = p

			} else if words[0] == "clusters" {
				mc[words[1]] = words[3:]
				lg.Debug(words[1])
				lg.Debug(fmt.Sprintf("%v", words[3:]))
			}
		}
	}
	config.Parameters = mp
	config.Clusters = mc

	lbclusters, err := loadClusters(&config, lg)
	if err != nil {
		fmt.Println("Error getting the clusters")
		return nil, nil, err
	}
	lg.Info("Clusters loaded")

	return &config, lbclusters, nil

}

func shouldUpdateDNS(config *Config, hostname string, lg *lbcluster.Log) bool {
	if hostname == config.Master {
		return true
	}
	masterHeartbeat := "I am sick"
	connectTimeout := (10 * time.Second)
	readWriteTimeout := (20 * time.Second)
	httpClient := lbcluster.NewTimeoutClient(connectTimeout, readWriteTimeout)
	response, err := httpClient.Get("http://" + config.Master + "/load-balancing/" + config.HeartbeatFile)
	if err != nil {
		lg.Warning(fmt.Sprintf("problem fetching heartbeat file from the primary master %v: %v", config.Master, err))
		return true
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		lg.Warning(fmt.Sprintf("%s", err))
	}
	lg.Debug(fmt.Sprintf("%s", contents))
	masterHeartbeat = strings.TrimSpace(string(contents))
	lg.Info("primary master heartbeat: " + masterHeartbeat)
	r, _ := regexp.Compile(config.Master + ` : (\d+) : I am alive`)
	if r.MatchString(masterHeartbeat) {
		matches := r.FindStringSubmatch(masterHeartbeat)
		lg.Debug(fmt.Sprintf(matches[1]))
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

func updateHeartbeat(config *Config, hostname string, lg *lbcluster.Log) error {
	if hostname != config.Master {
		return nil
	}
	heartbeatFile := config.HeartbeatPath + "/" + config.HeartbeatFile + "temp"
	heartbeatFileReal := config.HeartbeatPath + "/" + config.HeartbeatFile

	config.HeartbeatMu.Lock()
	defer config.HeartbeatMu.Unlock()

	f, err := os.OpenFile(heartbeatFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		lg.Error(fmt.Sprintf("can not open %v for writing: %v", heartbeatFile, err))
		return err
	}
	now := time.Now()
	secs := now.Unix()
	_, err = fmt.Fprintf(f, "%v : %v : I am alive\n", hostname, secs)
	lg.Info("updating: heartbeat file " + heartbeatFile)
	if err != nil {
		lg.Info(fmt.Sprintf("can not write to %v: %v", heartbeatFile, err))
	}
	f.Close()
	if err = os.Rename(heartbeatFile, heartbeatFileReal); err != nil {
		lg.Error(fmt.Sprintf("can not rename %v to %v: %v", heartbeatFile, heartbeatFileReal, err))
		return err
	}
	return nil
}

func installSignalHandler(sighup, sigterm *bool, lg *lbcluster.Log) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			// Block until a signal is received.
			sig := <-c
			lg.Info(fmt.Sprintf("\nGiven signal: %v\n", sig))
			switch sig {
			case syscall.SIGHUP:
				*sighup = true
			case syscall.SIGTERM:
				*sigterm = true
			}
		}
	}()
}

/* Using this one (instead of fsnotify)
to check also if the file has been moved*/
func watchFile(filePath string, chanModified chan int) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			chanModified <- 1
			initialStat = stat
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func sleep(seconds time.Duration, chanModified chan int) error {
	for {
		chanModified <- 2
		time.Sleep(seconds * time.Second)
	}
	return nil
}

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("This is a proof of concept golbd version %s \n", "0.001")
		os.Exit(0)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	log, e := syslog.New(syslog.LOG_NOTICE, "lbd")
	lg := lbcluster.Log{Writer: *log, Syslog: false, Stdout: *stdoutFlag, Debugflag: *debugFlag, TofilePath: *logFileFlag}
	if e == nil {
		lg.Info("Starting lbd")
	}

	//	var sig_hup, sig_term bool
	// installSignalHandler(&sig_hup, &sig_term, &lg)

	config, lbclusters, err := loadConfig(*configFileFlag, &lg)
	if err != nil {
		lg.Warning("loadConfig Error: ")
		lg.Warning(err.Error())
		os.Exit(1)
	}
	lg.Info("Clusters loaded")

	doneChan := make(chan int)
	go watchFile(*configFileFlag, doneChan)
	go sleep(10, doneChan)

	for {
		myValue := <-doneChan
		if myValue == 1 {
			lg.Info("Config Changed")
			config, lbclusters, err = loadConfig(*configFileFlag, &lg)
			if err != nil {
				lg.Error(fmt.Sprintf("Error getting the clusters (something wrong in %v", configFileFlag))
			}
		} else if myValue == 2 {
			checkAliases(config, lg, lbclusters)
		} else {
			lg.Error("Got an unexpected value")
		}
	}
	lg.Error("The lbd is not supposed to stop")

}
func checkAliases(config *Config, lg lbcluster.Log, lbclusters []lbcluster.LBCluster) {
	hostname, e := os.Hostname()
	if e == nil {
		lg.Info("Hostname: " + hostname)
	}

	//var wg sync.WaitGroup
	updateDNS := true
	lg.Info("Checking if any of the " + strconv.Itoa(len(lbclusters)) + " clusters needs updating")
	hostsToCheck := make(map[string]lbhost.LBHost)
	var clustersToUpdate []*lbcluster.LBCluster
	/* First, let's identify the hosts that have to be checked */
	for i := range lbclusters {
		pc := &lbclusters[i]
		pc.Write_to_log("DEBUG", "DO WE HAVE TO UPDATE?")
		if pc.Time_to_refresh() {
			pc.Write_to_log("INFO", "Time to refresh the cluster")
			pc.Get_list_hosts(hostsToCheck)
			clustersToUpdate = append(clustersToUpdate, pc)
		}
	}
	if len(hostsToCheck) != 0 {
		myChannel := make(chan lbhost.LBHost)
		/* Now, let's go through the hosts, issuing the snmp call */
		for _, hostValue := range hostsToCheck {
			go func(myHost lbhost.LBHost) {
				myHost.Snmp_req()
				myChannel <- myHost
			}(hostValue)
		}
		lg.Debug("Let's start gathering the results")
		for i := 0; i < len(hostsToCheck); i++ {
			myNewHost := <-myChannel
			hostsToCheck[myNewHost.Host_name] = myNewHost
		}

		lg.Debug("All the hosts have been tested")

		updateDNS = shouldUpdateDNS(config, hostname, &lg)

		/* Finally, let's go through the aliases, selecting the best hosts*/
		for _, pc := range clustersToUpdate {
			pc.Write_to_log("DEBUG", "READY TO UPDATE THE CLUSTER")
			pc.Find_best_hosts(hostsToCheck)
			if updateDNS {
				pc.Write_to_log("DEBUG", "Should update dns is true")
				pc.Refresh_dns(config.DNSManager, config.TsigKeyPrefix, config.TsigInternalKey, config.TsigExternalKey, hostsToCheck)
			} else {
				pc.Write_to_log("DEBUG", "should_update_dns false")
			}
		}
	}

	if updateDNS {
		updateHeartbeat(config, hostname, &lg)
	}

	lg.Debug("iteration done!")
}
