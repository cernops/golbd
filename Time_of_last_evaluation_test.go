package main

import (
	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"testing"
	"time"
)

func TestTimeOfLastEvaluation(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

	c := lbcluster.LBCluster{Cluster_name: "test01.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
		Host_metric_table:      map[string]int{"lxplus142.cern.ch": 100000, "lxplus177.cern.ch": 100000},
		Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_hosts:      []string{"unknown"},
		Previous_best_hosts:     []string{"unknown"},
		Previous_best_hosts_dns: []string{"unknown"},
		Slog:                    &lg,
		Current_index:           0}

	c.Time_of_last_evaluation = time.Now().Add(time.Duration(-c.Parameters.Polling_interval+2) * time.Second)
	if c.Time_to_refresh() {
		t.Errorf("e.Time_of_last_evaluation: got\n%v\nexpected\n%v", true, false)
	}
	c.Time_of_last_evaluation = time.Now().Add(time.Duration(-c.Parameters.Polling_interval-2) * time.Second)
	if !c.Time_to_refresh() {
		t.Errorf("e.Time_of_last_evaluation: got\n%v\nexpected\n%v", false, true)
	}
}
