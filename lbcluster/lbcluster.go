package lbcluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

const WorstValue int = 99999

type LBCluster struct {
	Cluster_name            string
	Loadbalancing_username  string
	Loadbalancing_password  string
	Host_metric_table       map[string]int
	Parameters              Params
	Time_of_last_evaluation time.Time
	Current_best_hosts      []string
	Previous_best_hosts     []string
	Statistics_filename     string
	Per_cluster_filename    string
	Current_index           int
}

type Params struct {
	Behaviour        string
	Best_hosts       int
	External         bool
	Metric           string
	Polling_interval int
	Statistics       string
}

func (self LBCluster) Time_to_refresh() bool {
	if self.Time_of_last_evaluation.IsZero() {
		return true
	} else {
		return self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval) * time.Second).After(time.Now())
	}
}

func (self LBCluster) find_best_hosts() {
	self.Previous_best_hosts = self.Current_best_hosts
	self.Evaluate_hosts()
}

func (self LBCluster) Evaluate_hosts() {
	for h := range self.Host_metric_table {
		fmt.Println("contacting cluster: " + self.Cluster_name + " node: " + h)
		metric, host, logm := self.snmp_req(h)
		fmt.Printf("%v\n%v %v\n", logm, host, metric)
	}
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

func (self LBCluster) snmp_req(host string) (int, string, string) {
	metric := -100
	logmessage := ""
	if self.Parameters.Metric == "cmsweb" {
		connectTimeout := (20 * time.Second)
		readWriteTimeout := (40 * time.Second)
		httpClient := NewTimeoutClient(connectTimeout, readWriteTimeout)
		response, err := httpClient.Get("http://woger-direct.cern.ch:9098/roger/v1/state/" + host)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		} else {
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
			}
			var dat map[string]interface{}
			if err := json.Unmarshal([]byte(contents), &dat); err != nil {
				panic(err)
			}
			if str, ok := dat["appstate"].(string); ok {
				if str != "production" {
					metric = -99
					logmessage = fmt.Sprintf("cluster: %s node: %s - %s - setting reply %v", self.Cluster_name, host, str, metric)
					return metric, host, logmessage
				}
			} else {
				fmt.Printf("dat[\"appstate\"] not a string for node %s", host)
			}

		}
		metric = WorstValue - 1
	}
	return metric, host, logmessage
}
