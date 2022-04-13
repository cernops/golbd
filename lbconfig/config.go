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

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gopkg.in/yaml.v3"
)

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
	ConfigFile      string
	Clusters        map[string][]string
	Parameters      map[string]lbcluster.Params
}

func LoadConfig(configFile string, lg *lbcluster.Log) (*Config, []lbcluster.LBCluster, error) {
	var configFunc func(configFile string, lg *lbcluster.Log) (*Config, []lbcluster.LBCluster, error)

	if strings.HasSuffix(configFile, ".yaml") {
		configFunc = loadConfigYaml
	} else {
		configFunc = loadConfigOriginal
	}

	return configFunc(configFile, lg)
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

//LoadClusters checks the syntax of the clusters defined in the configuration file
func LoadClusters(config *Config, lg *lbcluster.Log) ([]lbcluster.LBCluster, error) {
	var lbc lbcluster.LBCluster
	var lbcs []lbcluster.LBCluster

	for k, v := range config.Clusters {
		if len(v) == 0 {
			lg.Warning("cluster: " + k + " ignored as it has no members defined in the configuration file " + config.ConfigFile)
			continue
		}
		if par, ok := config.Parameters[k]; ok {
			lbc = lbcluster.LBCluster{Cluster_name: k, Loadbalancing_username: "loadbalancing",
				Loadbalancing_password: config.SnmpPassword, Parameters: par,
				Current_best_ips:      []net.IP{},
				Previous_best_ips_dns: []net.IP{},
				Slog:                  lg}
			hm := make(map[string]lbcluster.Node)
			for _, h := range v {
				hm[h] = lbcluster.Node{Load: 100000, IPs: []net.IP{}}
			}
			lbc.Host_metric_table = hm
			lbcs = append(lbcs, lbc)
			lbc.Write_to_log("INFO", "(re-)loaded cluster ")

		} else {
			lg.Warning("cluster: " + k + " missing parameters for cluster; ignoring the cluster, please check the configuration file " + config.ConfigFile)
		}
	}

	return lbcs, nil

}

//LoadConfigYaml reads a YAML configuration file and returns a struct with the config
func loadConfigYaml(configFile string, lg *lbcluster.Log) (*Config, []lbcluster.LBCluster, error) {
	var config Config

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, err
	}

	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return nil, nil, err
	}

	config.ConfigFile = configFile

	lbclusters, err := LoadClusters(&config, lg)
	if err != nil {
		fmt.Println("Error getting the clusters")
		return nil, nil, err
	}
	lg.Info("Clusters loaded")

	return &config, lbclusters, nil
}

//LoadConfig reads a configuration file and returns a struct with the config
func loadConfigOriginal(configFile string, lg *lbcluster.Log) (*Config, []lbcluster.LBCluster, error) {
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
	config.ConfigFile = configFile

	lbclusters, err := LoadClusters(&config, lg)
	if err != nil {
		fmt.Println("Error getting the clusters")
		return nil, nil, err
	}
	lg.Info("Clusters loaded")

	return &config, lbclusters, nil
}
