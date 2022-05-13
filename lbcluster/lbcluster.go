package lbcluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
	"lb-experts/golbd/model"
)

//WorstValue worst possible load
const WorstValue int = 99999

//LBCluster struct of an lbcluster alias
type LBCluster struct {
	ClusterConfig           model.CluserConfig
	Host_metric_table       map[string]Node
	Parameters              Params
	Time_of_last_evaluation time.Time
	Current_best_ips        []net.IP
	Previous_best_ips_dns   []net.IP
	Current_index           int
	Slog                    logger.Logger
}

//Params of the alias
type Params struct {
	Behaviour        string
	Best_hosts       int
	External         bool
	Metric           string
	Polling_interval int
	Statistics       string
	Ttl              int
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) error {
	if n < 0 {
		return fmt.Errorf("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
	return nil
}

//Node Struct to keep the ips and load of a node for an alias
type Node struct {
	Load     int
	IPs      []net.IP
	HostName string
}

//NodeList struct for the list
type NodeList []Node

func (p NodeList) Len() int           { return len(p) }
func (p NodeList) Less(i, j int) bool { return p[i].Load < p[j].Load }
func (p NodeList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

//Time_to_refresh Checks if the cluster needs refreshing
func (lbc *LBCluster) Time_to_refresh() bool {
	return lbc.Time_of_last_evaluation.Add(time.Duration(lbc.Parameters.Polling_interval) * time.Second).Before(time.Now())
}

//GetHostList Get the hosts for an alias
func (lbc *LBCluster) GetHostList() map[string]lbhost.Host {
	hostMap := make(map[string]lbhost.Host)
	lbc.Slog.Debug("Getting the list of hosts for the alias")
	for host := range lbc.Host_metric_table {
		myHost, ok := hostMap[host]
		if ok {
			clusterConfig := myHost.GetClusterConfig()
			clusterConfig.Cluster_name = clusterConfig.Cluster_name + "," + clusterConfig.Cluster_name
		} else {
			myHost = lbhost.NewLBHost(lbc.ClusterConfig, lbc.Slog)
		}
		hostMap[host] = myHost
	}
	return hostMap
}

func (lbc *LBCluster) concatenateNodes(myNodes []Node) string {
	nodes := make([]string, 0, len(myNodes))
	for _, node := range myNodes {
		nodes = append(nodes, lbc.concatenateIps(node.IPs))
	}
	return strings.Join(nodes, " ")
}

func (lbc *LBCluster) concatenateIps(myIps []net.IP) string {
	ip_string := make([]string, 0, len(myIps))

	for _, ip := range myIps {
		ip_string = append(ip_string, ip.String())
	}

	sort.Strings(ip_string)
	return strings.Join(ip_string, " ")
}

//Find_best_hosts Looks for the best hosts for a cluster
func (lbc *LBCluster) FindBestHosts(hosts_to_check map[string]lbhost.Host) (bool, error) {

	lbc.EvaluateHosts(hosts_to_check)
	allMetrics := make(map[string]bool)
	allMetrics["minimum"] = true
	allMetrics["cmsfrontier"] = true
	allMetrics["minino"] = true

	_, ok := allMetrics[lbc.Parameters.Metric]
	if !ok {
		lbc.Slog.Error("wrong parameter(metric) in definition of cluster " + lbc.Parameters.Metric)
		return false, nil
	}
	lbc.Time_of_last_evaluation = time.Now()
	shouldApplyMetric, err := lbc.ApplyMetric(hosts_to_check)
	if err != nil {
		return false, err
	}
	if !shouldApplyMetric {
		return false, nil
	}
	nodes := lbc.concatenateIps(lbc.Current_best_ips)
	if len(lbc.Current_best_ips) == 0 {
		nodes = "NONE"
	}
	lbc.Slog.Info("best hosts are: " + nodes)
	return true, nil
}

// ApplyMetric This is the core of the lbcluster: based on the metrics, select the best hosts
func (lbc *LBCluster) ApplyMetric(hosts_to_check map[string]lbhost.Host) (bool, error) {
	lbc.Slog.Info("Got metric = " + lbc.Parameters.Metric)
	pl := make(NodeList, len(lbc.Host_metric_table))
	i := 0
	for _, v := range lbc.Host_metric_table {
		pl[i] = v
		i++
	}
	//Let's shuffle the hosts before sorting them, in case some hosts have the same value
	err := Shuffle(len(pl), func(i, j int) { pl[i], pl[j] = pl[j], pl[i] })
	if err != nil {
		return false, err
	}
	sort.Sort(pl)
	lbc.Slog.Debug(fmt.Sprintf("%v", pl))
	var sorted_host_list []Node
	var useful_host_list []Node
	for _, v := range pl {
		if (v.Load > 0) && (v.Load <= WorstValue) {
			useful_host_list = append(useful_host_list, v)
		}
		sorted_host_list = append(sorted_host_list, v)
	}
	lbc.Slog.Debug(fmt.Sprintf("%v", useful_host_list))
	useful_hosts := len(useful_host_list)
	listLength := len(pl)
	max := lbc.Parameters.Best_hosts
	if max == -1 {
		max = listLength
	}
	if max > listLength {
		lbc.Slog.Warning(fmt.Sprintf("impossible to return %v hosts from the list of %v hosts (%v). Check the configuration of cluster. Returning %v hosts.",
			max, listLength, lbc.concatenateNodes(sorted_host_list), listLength))
		max = listLength
	}
	lbc.Current_best_ips = []net.IP{}
	if listLength == 0 {
		lbc.Slog.Error("cluster has no hosts defined ! Check the configuration.")
	} else if useful_hosts == 0 {

		if lbc.Parameters.Metric == "minimum" {
			lbc.Slog.Warning(fmt.Sprintf("no usable hosts found for cluster! Returning random %v hosts.", max))
			//Get hosts with all IPs even when not OK for SNMP
			lbc.ReEvaluateHostsForMinimum(hosts_to_check)
			i := 0
			for _, v := range lbc.Host_metric_table {
				pl[i] = v
				i++
			}
			//Let's shuffle the hosts
			err := Shuffle(len(pl), func(i, j int) { pl[i], pl[j] = pl[j], pl[i] })
			if err != nil {
				return false, err
			}
			for i := 0; i < max; i++ {
				lbc.Current_best_ips = append(lbc.Current_best_ips, pl[i].IPs...)
			}
			lbc.Slog.Warning(fmt.Sprintf("We have put random hosts behind the alias: %v", lbc.Current_best_ips))

		} else if (lbc.Parameters.Metric == "minino") || (lbc.Parameters.Metric == "cmsweb") {
			lbc.Slog.Warning("no usable hosts found for cluster! Returning no hosts.")
		} else if lbc.Parameters.Metric == "cmsfrontier" {
			lbc.Slog.Warning("no usable hosts found for cluster! Skipping the DNS update")
			return false, nil
		}
	} else {
		if useful_hosts < max {
			lbc.Slog.Warning(fmt.Sprintf("only %v useable hosts found in cluster", useful_hosts))
			max = useful_hosts
		}
		for i := 0; i < max; i++ {
			lbc.Current_best_ips = append(lbc.Current_best_ips, useful_host_list[i].IPs...)
		}
	}

	return true, nil
}

//NewTimeoutClient checks the timeout
/* The following functions are for the roger state and its timeout */
func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func (lbc *LBCluster) checkRogerState(host string) string {

	logmessage := ""

	connectTimeout := (10 * time.Second)
	readWriteTimeout := (20 * time.Second)
	httpClient := NewTimeoutClient(connectTimeout, readWriteTimeout)
	response, err := httpClient.Get("http://woger-direct.cern.ch:9098/roger/v1/state/" + host)
	if err != nil {
		logmessage = logmessage + fmt.Sprintf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logmessage = logmessage + fmt.Sprintf("%s", err)
		}
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(contents), &dat); err != nil {
			logmessage = logmessage + " - " + fmt.Sprintf("%s", host)
			logmessage = logmessage + " - " + fmt.Sprintf("%v", response.Body)
			logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
		}
		if str, ok := dat["appstate"].(string); ok {
			if str != "production" {
				return fmt.Sprintf("node: %s - %s - setting reply -99", host, str)
			}
		} else {
			logmessage = logmessage + fmt.Sprintf("dat[\"appstate\"] not a string for node %s", host)
		}
	}
	return logmessage

}

//EvaluateHosts gets the load from the all the nodes
func (lbc *LBCluster) EvaluateHosts(hostsToCheck map[string]lbhost.Host) {
	var nodeChan = make(chan Node)
	defer close(nodeChan)
	var wg sync.WaitGroup
	for currentHost := range lbc.Host_metric_table {
		wg.Add(1)
		go func(selectedHost string) {
			host := hostsToCheck[selectedHost]
			ips, err := host.GetWorkingIPs()
			if err != nil {
				ips, err = host.GetIps()
			}
			nodeChan <- Node{host.GetLoadForAlias(lbc.ClusterConfig.Cluster_name), ips, selectedHost}
		}(currentHost)
	}
	go func() {
		for nodeData := range nodeChan {
			lbc.Host_metric_table[nodeData.HostName] = nodeData
			lbc.Slog.Debug(fmt.Sprintf("node: %s It has a load of %d", nodeData.HostName, lbc.Host_metric_table[nodeData.HostName].Load))
			wg.Done()
		}
	}()
	wg.Wait()
}

//ReEvaluateHostsForMinimum gets the load from the all the nodes for Minimum metric policy
func (lbc *LBCluster) ReEvaluateHostsForMinimum(hostsToCheck map[string]lbhost.Host) {

	for currenthost := range lbc.Host_metric_table {
		host := hostsToCheck[currenthost]
		ips, err := host.GetAllIPs()
		if err != nil {
			ips, err = host.GetIps()
		}
		lbc.Host_metric_table[currenthost] = Node{host.GetLoadForAlias(lbc.ClusterConfig.Cluster_name), ips, host.GetName()}
		lbc.Slog.Debug(fmt.Sprintf("node: %s It has a load of %d", currenthost, lbc.Host_metric_table[currenthost].Load))
	}
}
