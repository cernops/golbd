package main_test

import (
	"reflect"
	"testing"
	"time"
)

func TestFindBestHosts(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	hosts_to_check := getHostsToCheck(c)

	expected_host_metric_table := map[string]int{"monit-kafkax-17be060b0d.cern.ch": 816, "lxplus132.cern.ch": 2, "lxplus041.cern.ch": 3, "lxplus130.cern.ch": 27}
	expected_previous_best_hosts := c.Current_best_hosts
	expected_current_best_hosts := []string{"lxplus041.cern.ch", "lxplus132.cern.ch"}

	c.Find_best_hosts(hosts_to_check)
	if !reflect.DeepEqual(c.Host_metric_table, expected_host_metric_table) {
		t.Errorf("e.Find_best_hosts: c.Host_metric_table: got\n%v\nexpected\n%v", c.Host_metric_table, expected_host_metric_table)
	}
	if !reflect.DeepEqual(c.Previous_best_hosts, expected_previous_best_hosts) {
		t.Errorf("e.Find_best_hosts: c.Previous_best_hosts: got\n%v\nexpected\n%v", c.Previous_best_hosts, expected_previous_best_hosts)
	}
	if !reflect.DeepEqual(c.Current_best_hosts, expected_current_best_hosts) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_hosts, expected_current_best_hosts)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(time.Now()) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\ncurrent time\n%v", c.Time_of_last_evaluation, time.Now())
	}
}
