package main_test

import (
	"reflect"
	"testing"
)

func TestEvaluateMetric(t *testing.T) {
	c := getTestClusterVariableMetric()
	expected_host_metric_table := c.Host_metric_table
	expected_previous_best_hosts := c.Previous_best_hosts
	expected_current_best_hosts := []string{"lxplus041.cern.ch", "lxplus132.cern.ch"}
	expected_time_of_last_evaluation := c.Time_of_last_evaluation

	c.Apply_metric()
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
