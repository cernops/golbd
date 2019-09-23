package main_test

import (
	"reflect"
	"testing"
)

func TestEvaluateHosts(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	hostsToCheck := getHostsToCheck(c)

	expectedHostMetricTable := map[string]int{"monit-kafkax-17be060b0d.cern.ch": 816, "lxplus132.cern.ch": 2, "lxplus041.cern.ch": 3, "lxplus130.cern.ch": 27}
	expectedPreviousBestHosts := c.Previous_best_hosts
	expectedCurrentBestHosts := c.Current_best_hosts
	expectedTimeOfLastEvaluation := c.Time_of_last_evaluation

	c.Evaluate_hosts(hostsToCheck)
	if !reflect.DeepEqual(c.Host_metric_table, expectedHostMetricTable) {
		t.Errorf("e.evaluate_hosts: c.Host_metric_table: got\n%v\nexpected\n%v", c.Host_metric_table, expectedHostMetricTable)
	}
	if !reflect.DeepEqual(c.Previous_best_hosts, expectedPreviousBestHosts) {
		t.Errorf("e.evaluate_hosts: c.Previous_best_hosts: got\n%v\nexpected\n%v", c.Previous_best_hosts, expectedPreviousBestHosts)
	}
	if !reflect.DeepEqual(c.Current_best_hosts, expectedCurrentBestHosts) {
		t.Errorf("e.evaluate_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_hosts, expectedCurrentBestHosts)
	}
	if !reflect.DeepEqual(c.Time_of_last_evaluation, expectedTimeOfLastEvaluation) {
		t.Errorf("e.evaluate_hosts: c.Time_of_last_evaluation: got\n%v\nexpected\n%v", c.Time_of_last_evaluation, expectedTimeOfLastEvaluation)
	}
}
