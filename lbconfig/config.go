package lbconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"lb-experts/golbd/lbcluster"
)

const (
	DefaultLoadBalancerConfig = "loadbalancing"
)

type Config interface {
	GetMasterHost() string
	GetHeartBeatFileName() string
	GetHeartBeatDirPath() string
	GetDNSManager() string
	GetTSIGKeyPrefix() string
	GetTSIGInternalKey() string
	GetTSIGExternalKey() string
	LockHeartBeatMutex()
	UnlockHeartBeatMutex()
	WatchFileChange(controlChan <-chan bool, waitGroup *sync.WaitGroup) <-chan ConfigFileChangeSignal
	Load() (*LBConfig, []lbcluster.LBCluster, error)
}

// Config this is the configuration of the lbd
type LBConfig struct {
	Master          string
	HeartbeatFile   string
	HeartbeatPath   string
	HeartbeatMu     sync.Mutex
	TsigKeyPrefix   string
	TsigInternalKey string
	TsigExternalKey string
	SnmpPassword    string
	DNSManager      string
	configFilePath  string
	lbLog           lbcluster.Logger
	Clusters        map[string][]string
	Parameters      map[string]lbcluster.Params
}

type ConfigFileChangeSignal struct {
	readSignal bool
	readError  error
}

func (fs ConfigFileChangeSignal) IsErrorPresent() bool {
	return fs.readError != nil
}

// NewLoadBalancerConfig - instantiates a new load balancer config
func NewLoadBalancerConfig(configFilePath string, lbClusterLog lbcluster.Logger) Config {
	return &LBConfig{
		configFilePath: configFilePath,
		lbLog:          lbClusterLog,
	}
}

func (c *LBConfig) GetMasterHost() string {
	return c.Master
}

func (c *LBConfig) GetHeartBeatFileName() string {
	return c.HeartbeatFile
}

func (c *LBConfig) GetHeartBeatDirPath() string {
	return c.HeartbeatPath
}

func (c *LBConfig) GetDNSManager() string {
	return c.DNSManager
}

func (c *LBConfig) GetTSIGKeyPrefix() string {
	return c.TsigKeyPrefix
}

func (c *LBConfig) GetTSIGInternalKey() string {
	return c.TsigInternalKey
}

func (c *LBConfig) GetTSIGExternalKey() string {
	return c.TsigExternalKey
}

func (c *LBConfig) LockHeartBeatMutex() {
	c.HeartbeatMu.Lock()
}

func (c *LBConfig) UnlockHeartBeatMutex() {
	c.HeartbeatMu.Unlock()
}

func (c *LBConfig) WatchFileChange(controlChan <-chan bool, waitGroup *sync.WaitGroup) <-chan ConfigFileChangeSignal {
	fileWatcherChan := make(chan ConfigFileChangeSignal)
	waitGroup.Add(1)
	go func() {
		defer close(fileWatcherChan)
		defer waitGroup.Done()
		initialStat, err := os.Stat(c.configFilePath)
		if err != nil {
			fileWatcherChan <- ConfigFileChangeSignal{readError: err}
		}
		secondTicker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-secondTicker.C:
				stat, err := os.Stat(c.configFilePath)
				if err != nil {
					fileWatcherChan <- ConfigFileChangeSignal{readError: err}
					continue
				}
				if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
					fileWatcherChan <- ConfigFileChangeSignal{readSignal: true}
					initialStat = stat
				}
			case <-controlChan:
				return
			}
		}
	}()
	return fileWatcherChan
}

//Load reads a configuration file and returns a struct with the config
func (c *LBConfig) Load() (*LBConfig, []lbcluster.LBCluster, error) {
	var (
		config LBConfig
		p      lbcluster.Params
		mc     = make(map[string][]string)
		mp     = make(map[string]lbcluster.Params)
	)

	lines, err := readLines(c.configFilePath)
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
					c.lbLog.Warning(fmt.Sprintf("%v", err))
					os.Exit(1)
				}
				mp[words[1]] = p

			} else if words[0] == "clusters" {
				mc[words[1]] = words[3:]
				c.lbLog.Debug(words[1])
				c.lbLog.Debug(fmt.Sprintf("%v", words[3:]))
			}
		}
	}
	config.Parameters = mp
	config.Clusters = mc

	lbclusters, err := c.loadClusters()
	if err != nil {
		fmt.Println("Error getting the clusters")
		return nil, nil, err
	}
	c.lbLog.Info("Clusters loaded")

	return &config, lbclusters, nil

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

//loadClusters checks the syntax of the clusters defined in the configuration file
func (c *LBConfig) loadClusters() ([]lbcluster.LBCluster, error) {
	var lbc lbcluster.LBCluster
	var lbcs []lbcluster.LBCluster

	for k, v := range c.Clusters {
		if len(v) == 0 {
			c.lbLog.Warning("cluster: " + k + " ignored as it has no members defined in the configuration file " + c.configFilePath)
			continue
		}
		if par, ok := c.Parameters[k]; ok {
			lbcConfig := lbcluster.Config{
				Cluster_name:           k,
				Loadbalancing_username: DefaultLoadBalancerConfig,
				Loadbalancing_password: c.SnmpPassword,
			}
			lbc = lbcluster.LBCluster{
				ClusterConfig:         lbcConfig,
				Parameters:            par,
				Current_best_ips:      []net.IP{},
				Previous_best_ips_dns: []net.IP{},
				Slog:                  c.lbLog,
			}
			hm := make(map[string]lbcluster.Node)
			for _, h := range v {
				hm[h] = lbcluster.Node{Load: 100000, IPs: []net.IP{}}
			}
			lbc.Host_metric_table = hm
			lbcs = append(lbcs, lbc)
			lbc.Slog.Info("(re-)loaded cluster ")

		} else {
			c.lbLog.Warning("cluster: " + k + " missing parameters for cluster; ignoring the cluster, please check the configuration file " + c.configFilePath)
		}
	}

	return lbcs, nil

}
