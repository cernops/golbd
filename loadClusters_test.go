package main

import (
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
)

func TestLoadClusters(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

	config := Config{Master: "lbdxyz.cern.ch",
		HeartbeatFile: "heartbeat",
		HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
		//HeartbeatMu:     sync.Mutex{0, 0},
		TsigKeyPrefix:   "abcd-",
		TsigInternalKey: "xxx123==",
		TsigExternalKey: "yyy123==",
		SnmpPassword:    "zzz123",
		DnsManager:      "111.111.0.111",
		Clusters:        map[string][]string{"test01.cern.ch": []string{"lxplus142.cern.ch", "lxplus177.cern.ch"}, "test02.cern.ch": []string{"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus025.cern.ch"}},
		Parameters: map[string]lbcluster.Params{"test01.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 2,
			External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			"test02.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"}}}
	expected := []lbcluster.LBCluster{
		lbcluster.LBCluster{Cluster_name: "test01.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			Host_metric_table:      map[string]int{"lxplus142.cern.ch": 100000, "lxplus177.cern.ch": 100000},
			Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"unknown"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:          &lg,
			Current_index: 0},
		lbcluster.LBCluster{Cluster_name: "test02.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			Host_metric_table:      map[string]int{"lxplus013.cern.ch": 100000, "lxplus038.cern.ch": 100000, "lxplus025.cern.ch": 100000},
			Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_hosts:      []string{"unknown"},
			Previous_best_hosts:     []string{"unknown"},
			Previous_best_hosts_dns: []string{"unknown"},
			Slog:          &lg,
			Current_index: 0}}

	lbclusters := loadClusters(&config, &lg)
	// reflect.DeepEqual(lbclusters, expected) occassionally fails as the array order is not always the same
	// so comparing element par element
	i := 0
	for _, e := range expected {
		for _, c := range lbclusters {
			if c.Cluster_name == e.Cluster_name {
				if !reflect.DeepEqual(c, e) {
					t.Errorf("loadClusters: got\n%v\nexpected\n%v", lbclusters, expected)
				} else {
					i = i + 1
				}
				continue
			}
		}
	}
	if (i != len(expected)) || (i != len(lbclusters)) {
		t.Errorf("loadClusters: wrong number of clusters, got\n%v\nexpected\n%v", lbclusters, expected)

	}
}
