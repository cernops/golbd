package lbcluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"

	"gitlab.cern.ch/lb-experts/golbd/lbhost"

	"sort"
	"time"
)

//WorstValue worst possible load
const WorstValue int = 99999

//TIMEOUT snmp timeout
const TIMEOUT int = 10

//OID snmp object to get
const OID string = ".1.3.6.1.4.1.96.255.1"

//LBCluster struct of an lbcluster alias
type LBCluster struct {
	Cluster_name            string
	Loadbalancing_username  string
	Loadbalancing_password  string
	Host_metric_table       map[string]Node
	Parameters              Params
	Time_of_last_evaluation time.Time
	Current_best_ips        []net.IP
	Previous_best_ips_dns   []net.IP
	Current_index           int
	Slog                    *Log
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
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
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
}

//Node Struct to keep the ips and load of a node for an alias
type Node struct {
	Load int
	IPs  []net.IP
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

//Get_list_hosts Get the hosts for an alias
func (lbc *LBCluster) Get_list_hosts(current_list map[string]lbhost.LBHost) {
	lbc.Write_to_log("DEBUG", "Getting the list of hosts for the alias")
	for host := range lbc.Host_metric_table {
		myHost, ok := current_list[host]
		if ok {
			myHost.Cluster_name = myHost.Cluster_name + "," + lbc.Cluster_name
		} else {
			myHost = lbhost.LBHost{
				Cluster_name:           lbc.Cluster_name,
				Host_name:              host,
				Loadbalancing_username: lbc.Loadbalancing_username,
				Loadbalancing_password: lbc.Loadbalancing_password,
				LogFile:                lbc.Slog.TofilePath,
				Debugflag:              lbc.Slog.Debugflag,
			}
		}
		current_list[host] = myHost
	}
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
func (lbc *LBCluster) FindBestHosts(hosts_to_check map[string]lbhost.LBHost) {

	lbc.EvaluateHosts(hosts_to_check)
	allMetrics := make(map[string]bool)
	allMetrics["minimum"] = true
	allMetrics["cmsfrontier"] = true
	allMetrics["minino"] = true

	_, ok := allMetrics[lbc.Parameters.Metric]
	if !ok {
		lbc.Write_to_log("ERROR", "wrong parameter(metric) in definition of cluster "+lbc.Parameters.Metric)
		return
	}
	lbc.ApplyMetric()
	lbc.Time_of_last_evaluation = time.Now()
	nodes := lbc.concatenateIps(lbc.Current_best_ips)
	if len(lbc.Current_best_ips) == 0 {
		nodes = "NONE"
	}
	lbc.Write_to_log("INFO", "best hosts are: "+nodes)
}

// ApplyMetric This is the core of the lbcluster: based on the metrics, select the best hosts
func (lbc *LBCluster) ApplyMetric() {
	lbc.Write_to_log("INFO", "Got metric = "+lbc.Parameters.Metric)
	pl := make(NodeList, len(lbc.Host_metric_table))
	i := 0
	for _, v := range lbc.Host_metric_table {
		pl[i] = v
		i++
	}
	//Let's shuffle the hosts before sorting them, in case some hosts have the same value
	Shuffle(len(pl), func(i, j int) { pl[i], pl[j] = pl[j], pl[i] })
	sort.Sort(pl)
	lbc.Write_to_log("DEBUG", fmt.Sprintf("%v", pl))
	var sorted_host_list []Node
	var useful_host_list []Node
	for _, v := range pl {
		if (v.Load > 0) && (v.Load <= WorstValue) {
			useful_host_list = append(useful_host_list, v)
		}
		sorted_host_list = append(sorted_host_list, v)
	}
	lbc.Write_to_log("DEBUG", fmt.Sprintf("%v", useful_host_list))
	useful_hosts := len(useful_host_list)
	listLength := len(pl)
	max := lbc.Parameters.Best_hosts
	if max == -1 {
		max = listLength
	}
	if max > listLength {
		lbc.Write_to_log("WARNING", fmt.Sprintf("impossible to return %v hosts from the list of %v hosts (%v). Check the configuration of cluster. Returning %v hosts.", max, listLength, sorted_host_list, listLength))
		max = listLength
	}
	lbc.Current_best_ips = []net.IP{}
	if listLength == 0 {
		lbc.Write_to_log("ERROR", "cluster has no hosts defined ! Check the configuration.")
	} else if useful_hosts == 0 {
		putRandomHosts := false
		if lbc.Parameters.Metric == "minimum" {
			lbc.Write_to_log("WARNING", fmt.Sprintf("no usable hosts found for cluster! Returning random %v hosts.", max))
			putRandomHosts = true
		} else if (lbc.Parameters.Metric == "minino") || (lbc.Parameters.Metric == "cmsweb") {
			lbc.Write_to_log("WARNING", "no usable hosts found for cluster! Returning no hosts.")
		} else if lbc.Parameters.Metric == "cmsfrontier" {
			lbc.Write_to_log("WARNING", "no usable hosts found for cluster!, using the previous_best_hosts")
			if len(lbc.Previous_best_ips_dns) != 0 {
				//If there was something in the DNS, keep that
				lbc.Current_best_ips = lbc.Previous_best_ips_dns
			} else {
				//Otherwise, let's put random hosts
				putRandomHosts = true
			}
		}
		if putRandomHosts {
			Shuffle(len(sorted_host_list), func(i, j int) {
				sorted_host_list[i], sorted_host_list[j] = sorted_host_list[j], sorted_host_list[i]
			})
			for i := 0; i < max; i++ {
				lbc.Current_best_ips = append(lbc.Current_best_ips, sorted_host_list[i].IPs...)
			}
		}
	} else {
		if useful_hosts < max {
			lbc.Write_to_log("WARNING", fmt.Sprintf("only %v useable hosts found in cluster", useful_hosts))
			max = useful_hosts
		}
		for i := 0; i < max; i++ {
			lbc.Current_best_ips = append(lbc.Current_best_ips, useful_host_list[i].IPs...)
		}
	}
	sort.Slice(lbc.Current_best_ips, func(i, j int) bool {
		return bytes.Compare(lbc.Current_best_ips[i], lbc.Current_best_ips[j]) < 0
	})

	return
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
func (lbc *LBCluster) EvaluateHosts(hostsToCheck map[string]lbhost.LBHost) {

	for currenthost := range lbc.Host_metric_table {
		host := hostsToCheck[currenthost]
		ips, err := host.Get_working_IPs()
		if err != nil {
			ips, err = host.Get_Ips()
		}
		lbc.Host_metric_table[currenthost] = Node{host.Get_load_for_alias(lbc.Cluster_name), ips}
		lbc.Write_to_log("DEBUG", fmt.Sprintf("node: %s It has a load of %d", currenthost, lbc.Host_metric_table[currenthost].Load))
	}
}
