package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"

	"lb-experts/golbd/lbconfig"
)

var (
	// Version number
	// This should be overwritten with `go build -ldflags "-X main.Version='HELLO_THERE'"`
	Version = "head"
	// Release number
	// It should also be overwritten
	Release = "no_release"

	versionFlag    = flag.Bool("version", false, "print lbd version and exit")
	debugFlag      = flag.Bool("debug", false, "set lbd in debug mode")
	startFlag      = flag.Bool("start", false, "start lbd")
	stopFlag       = flag.Bool("stop", false, "stop lbd")
	updateFlag     = flag.Bool("update", false, "update lbd config")
	configFileFlag  = flag.String("config", "./load-balancing.conf", "specify configuration file path")
	logFileFlag    = flag.String("log", "./lbd.log", "specify log file path")
	stdoutFlag     = flag.Bool("stdout", false, "send log to stdtout")
)

const (
	itCSgroupDNSserver   string = "cfmgr.cern.ch"
	DefaultSleepDuration        = 10
	DefaultLbdTag               ="lbd"
	DefaultConnectionTimeout               =10 * time.Second
	DefaultReadTimeout               =20 * time.Second
)

type ConfigFileChangeSignal struct {
	readSignal bool
	readError error
}

func shouldUpdateDNS(config lbconfig.Config, hostname string, lg *lbcluster.Log) bool {
	if strings.EqualFold( hostname, config.GetMasterHost()) {
		return true
	}
	masterHeartbeat := "I am sick"
	httpClient := lbcluster.NewTimeoutClient(DefaultConnectionTimeout, DefaultReadTimeout)
	response, err := httpClient.Get("http://" + config.GetMasterHost() + "/load-balancing/" + config.HeartbeatFile)
	if err != nil {
		lg.Warning(fmt.Sprintf("problem fetching heartbeat file from the primary master %v: %v", config.GetMasterHost(), err))
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
	r, _ := regexp.Compile(config.GetMasterHost() + ` : (\d+) : I am alive`)
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

func updateHeartbeat(config lbconfig.Config, hostname string, lg *lbcluster.Log) error {
	if hostname != config.GetMasterHost() {
		return nil
	}
	heartbeatTempFilePath := config.GetHeartBeatDirPath() + "/" + config.GetHeartBeatFileName() + "temp"
	heartbeatFileRealFilePath := config.GetHeartBeatDirPath() + "/" + config.GetHeartBeatFileName()

	//todo: read from channel
	config.LockHeartBeatMutex()
	defer config.UnlockHeartBeatMutex()


	err := updateHeartBeatToFile(heartbeatTempFilePath, hostname, lg)
	if err != nil {
		return err
	}
	// todo: could the file be reused for any other use cases?
	if err = os.Rename(heartbeatTempFilePath, heartbeatFileRealFilePath); err != nil {
		lg.Error(fmt.Sprintf("can not rename %v to %v: %v", heartbeatTempFilePath, heartbeatFileRealFilePath, err))
		return err
	}
	return nil
}

func updateHeartBeatToFile(heartBeatFilePath string,  hostname string, lg *lbcluster.Log) error{
	secs := time.Now().Unix()
	f, err := os.OpenFile(heartBeatFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		lg.Error(fmt.Sprintf("can not open %v for writing: %v", heartBeatFilePath, err))
		return err
	}
	_, err = fmt.Fprintf(f, "%v : %v : I am alive\n", hostname, secs)
	lg.Info("updating: heartbeat file " + heartBeatFilePath)
	if err != nil {
		lg.Info(fmt.Sprintf("can not write to %v: %v", heartBeatFilePath, err))
	}
	return nil
}

func sleep(seconds time.Duration, controlChan <-chan bool, waitGroup *sync.WaitGroup) <-chan bool{
	sleepSignalChan := make(chan bool)
	waitGroup.Add(1)
	secondsTicker := time.NewTicker(seconds * time.Second)
	go func() {
		defer waitGroup.Done()
		for {
			select {
			case <- secondsTicker.C:
				sleepSignalChan <- true
				break
			case <- controlChan:
				return
			}
		}
	}()
	return sleepSignalChan
}

func main() {
	wg:= sync.WaitGroup{}
	log, e := syslog.New(syslog.LOG_NOTICE, DefaultLbdTag)
	lg := lbcluster.Log{SyslogWriter: log, Stdout: *stdoutFlag, Debugflag: *debugFlag, TofilePath: *logFileFlag}
	controlChan := make(chan bool)
	defer close(controlChan)
	defer wg.Done()
	defer lg.Error("The lbd is not supposed to stop")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("This is a proof of concept golbd version: %s-%s \n", Version, Release)
		os.Exit(0)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	if e != nil {
		fmt.Printf("Error getting a syslog instance %v\nThe service will only write to the logfile %v\n\n", e, *logFileFlag)
	}

	lg.Info("Starting lbd")
	lbConfig := lbconfig.NewLoadBalancerConfig(*configFileFlag, &lg)
	config, lbclusters, err := lbConfig.Load()
	if err != nil {
		lg.Warning("loadConfig Error: ")
		lg.Warning(err.Error())
		os.Exit(1)
	}
	lg.Info("Clusters loaded")

	fileChangeSignal := lbConfig.WatchFileChange(controlChan, &wg)
	intervalTickerSignal := sleep(DefaultSleepDuration, controlChan,&wg)
	for {
		select {
			case fileWatcherData := <-fileChangeSignal:
				if fileWatcherData.IsErrorPresent(){
					// stop all operations
					controlChan <- true
					return
				}
				lg.Info("Config Changed")
				config, lbclusters, err = lbConfig.Load()
				if err != nil {
					lg.Error(fmt.Sprintf("Error getting the clusters (something wrong in %v", configFileFlag))
				}
			case <-intervalTickerSignal:
				checkAliases(config, lg, lbclusters)
				break
		}
	}
}
// todo: add some tests
func checkAliases(config lbconfig.Config, lg lbcluster.Log, lbclusters []lbcluster.LBCluster) {
	hostCheckChannel := make(chan lbhost.LBHost)
	defer close(hostCheckChannel)

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
		currentCluster := &lbclusters[i]
		currentCluster.Write_to_log("DEBUG", "DO WE HAVE TO UPDATE?")
		if currentCluster.Time_to_refresh() {
			currentCluster.Write_to_log("INFO", "Time to refresh the cluster")
			currentCluster.Get_list_hosts(hostsToCheck)
			clustersToUpdate = append(clustersToUpdate, currentCluster)
		}
	}
	if len(hostsToCheck) > 0 {
		/* Now, let's go through the hosts, issuing the snmp call */
		for _, hostValue := range hostsToCheck {
			go func(myHost lbhost.LBHost) {
				myHost.Snmp_req()
				hostCheckChannel <- myHost
			}(hostValue)
		}
		lg.Debug("start gathering the results")
		for hostChanData := range hostCheckChannel {
			hostsToCheck[hostChanData.Host_name] = hostChanData
		}

		lg.Debug("All the hosts have been tested")

		updateDNS = shouldUpdateDNS(config, hostname, &lg)

		/* Finally, let's go through the aliases, selecting the best hosts*/
		//todo: try to update clusters in parallel
		for _, pc := range clustersToUpdate {
			pc.Write_to_log("DEBUG", "READY TO UPDATE THE CLUSTER")
			if pc.FindBestHosts(hostsToCheck) {
				if updateDNS {
					pc.Write_to_log("DEBUG", "Should update dns is true")
					// todo: try to implement retry mechanism
					pc.RefreshDNS(config.GetDNSManager(), config.GetTSIGKeyPrefix(), config.GetTSIGInternalKey(), config.GetTSIGExternalKey())
				} else {
					pc.Write_to_log("DEBUG", "should_update_dns false")
				}
			} else {
				pc.Write_to_log("DEBUG", "FindBestHosts false")
			}
		}
	}

	if updateDNS {
		updateHeartbeat(config, hostname, &lg)
	}

	lg.Debug("iteration done!")
}
