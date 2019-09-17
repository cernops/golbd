package lbcluster

import (
	"net"
	"reflect"
	"testing"
)

//TestGetStateDNS tests the function get_state_dns
func TestGetStateDNS(t *testing.T) {
	lg := Log{Syslog: false, Stdout: true, Debugflag: false}
	//DNS IP
	dnsManager := "137.138.16.5"
	//Empty slice for comparisson purposes
	var ipsEmpty []net.IP
	//Definition of expected hosts IP for aiermis, valid in the time when the test was written
	ExpectedIPAiermis := []net.IP{
		//IPv4 , for some unknown reason requires to be formatted in 4 Bytes (.To4())
		net.ParseIP("188.184.104.111"),
		net.ParseIP("2001:1458:d00:2d::100:58"),
	}

	//Non-existing clusters

	Clusters := []LBCluster{
		LBCluster{Cluster_name: "testme007.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			Host_metric_table:      map[string]int{"lxplus132.cern.ch": 100000, "lxplus041.cern.ch": 100000, "lxplus130.cern.ch": 100000, "monit-kafkax-17be060b0d.cern.ch": 100000},
			Parameters:             Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"unknown"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:                    &lg,
			Current_index:           0},

		LBCluster{Cluster_name: "testme007",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			Host_metric_table:      map[string]int{"lxplus132.cern.ch": 100000, "lxplus041.cern.ch": 100000, "lxplus130.cern.ch": 100000, "monit-kafkax-17be060b0d.cern.ch": 100000},
			Parameters:             Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"unknown"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:                    &lg,
			Current_index:           0},

		//Existing clusters

		LBCluster{Cluster_name: "kkouros.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			Host_metric_table:      map[string]int{"lxplus132.cern.ch": 100000, "lxplus041.cern.ch": 100000, "lxplus130.cern.ch": 100000, "monit-kafkax-17be060b0d.cern.ch": 100000},
			Parameters:             Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"unknown"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:                    &lg,
			Current_index:           0},

		LBCluster{Cluster_name: "aiermis.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			Host_metric_table:      map[string]int{"lxplus132.cern.ch": 100000, "lxplus041.cern.ch": 100000, "lxplus130.cern.ch": 100000, "monit-kafkax-17be060b0d.cern.ch": 100000},
			Parameters:             Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"ermistest1.cern.ch", "ermistest2.cern.ch"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:                    &lg,
			Current_index:           0},
	}
	//Expected response for every alias ( slice of IP and error message if any)
	expected := map[string][]interface{}{
		"testme007.cern.ch": {ipsEmpty, nil},
		"testme007":         {ipsEmpty, nil},
		"kkouros.cern.ch":   {ipsEmpty, nil},
		"aiermis.cern.ch":   {ExpectedIPAiermis, nil},
	}
	//receiving the output for every alias and storing the results into a map
	received := make(map[string][]interface{})
	for _, c := range Clusters {
		iprec, err := c.get_state_dns(dnsManager)

		received[c.Cluster_name] = []interface{}{iprec, err}
	}
	//DeepEqual comparison between the map with expected values and the one with the outputs
	for _, c := range Clusters {
		if !reflect.DeepEqual(received[c.Cluster_name], expected[c.Cluster_name]) {
			t.Errorf("\ngot\n%v\nexpected\n%v", received[c.Cluster_name], expected[c.Cluster_name])
		}
	}

}
