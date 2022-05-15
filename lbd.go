package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"lb-experts/golbd/metric"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbconfig"
	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
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
	configFileFlag = flag.String("config", "./load-balancing.[conf][yaml]", "specify configuration file path")
	logFileFlag    = flag.String("log", "./lbd.log", "specify log file path")
	stdoutFlag     = flag.Bool("stdout", false, "send log to stdtout")
)

const (
	shouldStartMetricServer  = false // server disabled by default
	itCSgroupDNSserver       = "cfmgr.cern.ch"
	DefaultSleepDuration     = 10
	DefaultLbdTag            = "lbd"
	DefaultConnectionTimeout = 10 * time.Second
	DefaultReadTimeout       = 20 * time.Second
)

type ConfigFileChangeSignal struct {
	readSignal bool
	readError  error
}

func shouldUpdateDNS(config lbconfig.Config, hostname string, lg logger.Logger) bool {
	if strings.EqualFold(hostname, config.GetMasterHost()) {
		return true
	}
	masterHeartbeat := "I am sick"
	httpClient := lbcluster.NewTimeoutClient(DefaultConnectionTimeout, DefaultReadTimeout)
	response, err := httpClient.Get("http://" + config.GetMasterHost() + "/load-balancing/" + config.GetHeartBeatFileName())
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

func updateHeartbeat(config lbconfig.Config, hostname string, lg logger.Logger) error {
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

func updateHeartBeatToFile(heartBeatFilePath string, hostname string, lg logger.Logger) error {
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

func sleep(seconds time.Duration, controlChan <-chan bool, waitGroup sync.WaitGroup) <-chan bool {
	sleepSignalChan := make(chan bool)
	waitGroup.Add(1)
	secondsTicker := time.NewTicker(seconds * time.Second)
	go func() {
		defer waitGroup.Done()
		for {
			select {
			case <-secondsTicker.C:
				sleepSignalChan <- true
				break
			case <-controlChan:
				return
			}
		}
	}()
	return sleepSignalChan
}

func main() {
	wg := sync.WaitGroup{}
	logger, err := logger.NewLoggerFactory(*logFileFlag)
	if err != nil {
		fmt.Printf("error during log initialization. error: %v", err)
		os.Exit(1)
	}

	if *stdoutFlag {
		logger.EnableWriteToSTd()
	}
	controlChan := make(chan bool)
	defer close(controlChan)
	defer wg.Done()
	defer logger.Error("The lbd is not supposed to stop")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("This is a proof of concept golbd version: %s-%s \n", Version, Release)
		os.Exit(0)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	logger.Info("Starting lbd")
	lbConfig := lbconfig.NewLoadBalancerConfig(*configFileFlag, logger)
	lbclusters, err := lbConfig.Load()
	if err != nil {
		logger.Warning("loadConfig Error: ")
		logger.Warning(err.Error())
		os.Exit(1)
	}
	logger.Info("Clusters loaded")

	fileChangeSignal := lbConfig.WatchFileChange(controlChan, wg)
	intervalTickerSignal := sleep(DefaultSleepDuration, controlChan, wg)
	if shouldStartMetricServer {
		go func() {
			err := metric.NewMetricServer(lbconfig.DefaultMetricsDirectoryPath)
			if err != nil {
				logger.Error(fmt.Sprintf("error while starting metric server . error: %v", err))
			}
		}()
	}

	for {
		select {
		case fileWatcherData := <-fileChangeSignal:
			if fileWatcherData.IsErrorPresent() {
				// stop all operations
				controlChan <- true
				return
			}
			logger.Info("ClusterConfig Changed")
			lbclusters, err = lbConfig.Load()
			if err != nil {
				logger.Error(fmt.Sprintf("Error getting the clusters (something wrong in %v", configFileFlag))
			}
		case <-intervalTickerSignal:
			checkAliases(lbConfig, logger, lbclusters)
			break
		}
	}
}

func checkAliases(config lbconfig.Config, lg logger.Logger, lbclusters []lbcluster.LBCluster) {
	hostCheckChannel := make(chan lbhost.Host)
	defer close(hostCheckChannel)

	hostname, e := os.Hostname()
	if e == nil {
		lg.Info("Hostname: " + hostname)
	}

	//var wg sync.WaitGroup
	updateDNS := true
	lg.Info("Checking if any of the " + strconv.Itoa(len(lbclusters)) + " clusters needs updating")
	var clustersToUpdate []*lbcluster.LBCluster
	hostsToCheck := make(map[string]lbhost.Host)
	/* First, let's identify the hosts that have to be checked */
	for i := range lbclusters {
		currentCluster := &lbclusters[i]
		lg.Debug("DO WE HAVE TO UPDATE?")
		if currentCluster.Time_to_refresh() {
			lg.Info("Time to refresh the cluster")
			currentCluster.GetHostList(hostsToCheck)
			clustersToUpdate = append(clustersToUpdate, currentCluster)
		}
	}
	if len(hostsToCheck) > 0 {
		/* Now, let's go through the hosts, issuing the snmp call */
		for _, hostValue := range hostsToCheck {
			go func(myHost lbhost.Host) {
				myHost.SNMPDiscovery()
				hostCheckChannel <- myHost
			}(hostValue)
		}
		lg.Debug("start gathering the results")
		for hostChanData := range hostCheckChannel {
			hostsToCheck[hostChanData.GetName()] = hostChanData
		}

		lg.Debug("All the hosts have been tested")

		updateDNS = shouldUpdateDNS(config, hostname, lg)

		/* Finally, let's go through the aliases, selecting the best hosts*/
		for _, pc := range clustersToUpdate {
			lg.Debug("READY TO UPDATE THE CLUSTER")
			isDNSUpdateValid, err := pc.FindBestHosts(hostsToCheck)
			if err != nil {
				log.Fatalf("Error while finding best hosts. error:%v", err)
			}
			if isDNSUpdateValid {
				if updateDNS {
					lg.Debug("Should update dns is true")
					// todo: try to implement retry mechanismlbcluster/lbcluster_dns.go
					pc.RefreshDNS(config.GetDNSManager(), config.GetTSIGKeyPrefix(), config.GetTSIGInternalKey(), config.GetTSIGExternalKey())
				} else {
					lg.Debug("should_update_dns false")
				}
			} else {
				lg.Debug("FindBestHosts false")
			}
		}
	}

	if updateDNS {
		updateHeartbeat(config, hostname, lg)
	}

	lg.Debug("iteration done!")
}
