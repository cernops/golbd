package lbcluster

import (
	"reflect"
	"testing"
)

func TestEvaluateMetric(t *testing.T) {
	lg := Log{Syslog: false, Stdout: true, Debugflag: false}

	c := LBCluster{Cluster_name: "test01.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "OidLvSu8amNMRz258Udy74tO60p47n0RA4RzaT3j2hhnJkEQg9",
		Host_metric_table:      map[string]int{"monit-kafkax-17be060b0d.cern.ch": 816, "lxplus132.cern.ch": 2, "lxplus041.cern.ch": 3, "lxplus130.cern.ch": 27},
		Parameters:             Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_hosts:      []string{"unknown"},
		Previous_best_hosts:     []string{"unknown"},
		Previous_best_hosts_dns: []string{"unknown"},
		Slog:                    &lg,
		Current_index:           0,
	}
	expected_host_metric_table := c.Host_metric_table
	expected_previous_best_hosts := c.Previous_best_hosts
	expected_current_best_hosts := []string{"lxplus041.cern.ch", "lxplus132.cern.ch"}
	expected_time_of_last_evaluation := c.Time_of_last_evaluation

	c.apply_metric()
	if !reflect.DeepEqual(c.Host_metric_table, expected_host_metric_table) {
		t.Errorf("e.apply_metric: c.Host_metric_table: got\n%v\nexpected\n%v", c.Host_metric_table, expected_host_metric_table)
	}
	if !reflect.DeepEqual(c.Previous_best_hosts, expected_previous_best_hosts) {
		t.Errorf("e.apply_metric: c.Previous_best_hosts: got\n%v\nexpected\n%v", c.Previous_best_hosts, expected_previous_best_hosts)
	}
	if !reflect.DeepEqual(c.Current_best_hosts, expected_current_best_hosts) {
		t.Errorf("e.apply_metric: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_hosts, expected_current_best_hosts)
	}
	if !reflect.DeepEqual(c.Time_of_last_evaluation, expected_time_of_last_evaluation) {
		t.Errorf("e.apply_metric: c.Time_of_last_evaluation: got\n%v\nexpected\n%v", c.Time_of_last_evaluation, expected_time_of_last_evaluation)
	}

}
