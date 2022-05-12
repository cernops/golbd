package main_test

import (
	"net"
	"reflect"
	"testing"
	"time"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
)

func getExpectedHostMetric() map[string]lbcluster.Node {
	return map[string]lbcluster.Node{
		"monit-kafkax-17be060b0d.cern.ch": lbcluster.Node{Load: 816, IPs: []net.IP{net.ParseIP("188.184.108.100")}},
		"lxplus132.cern.ch":               lbcluster.Node{Load: 2, IPs: []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"), net.ParseIP("188.184.108.98")}},
		"lxplus041.cern.ch":               lbcluster.Node{Load: 3, IPs: []net.IP{net.ParseIP("2001:1458:d00:32::100:51"), net.ParseIP("188.184.116.81")}},
		"lxplus130.cern.ch":               lbcluster.Node{Load: 27, IPs: []net.IP{net.ParseIP("188.184.108.100")}},
		"lxplus133.subdo.cern.ch":         lbcluster.Node{Load: 27, IPs: []net.IP{net.ParseIP("188.184.108.101")}}}
}

func TestFindBestHosts(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	hosts_to_check := getHostsToCheck(c)

	expected_host_metric_table := getExpectedHostMetric()

	expected_current_best_ips := []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"), net.ParseIP("188.184.108.98"), net.ParseIP("2001:1458:d00:32::100:51"), net.ParseIP("188.184.116.81")}
	isDNSUpdateValid, err := c.FindBestHosts(hosts_to_check)
	if err != nil {
		t.Errorf("Error while finding best hosts. error:%v", err)
	}
	if !isDNSUpdateValid {
		t.Errorf("e.Find_best_hosts: returned false, expected true")
	}
	if !reflect.DeepEqual(c.Host_metric_table, expected_host_metric_table) {
		t.Errorf("e.Find_best_hosts: c.Host_metric_table: got\n%v\nexpected\n%v", c.Host_metric_table, expected_host_metric_table)
	}
	if !reflect.DeepEqual(c.Current_best_ips, expected_current_best_ips) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_ips, expected_current_best_ips)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(time.Now()) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\ncurrent time\n%v", c.Time_of_last_evaluation, time.Now())
	}
}

func TestFindBestHostsNoValidHostCmsfrontier(t *testing.T) {

	c := getTestCluster("testbad.cern.ch")

	c.Parameters.Metric = "cmsfrontier"

	bad_hosts_to_check := getBadHostsToCheck(c)

	expected_current_best_ips := []net.IP{}

	expected_time_of_last_evaluation := c.Time_of_last_evaluation
	isDNSUpdateValid, err := c.FindBestHosts(bad_hosts_to_check)
	if err != nil {
		t.Errorf("Error while finding best hosts. error:%v", err)
	}
	if isDNSUpdateValid {
		t.Errorf("e.Find_best_hosts: returned true, expected false")
	}
	if !reflect.DeepEqual(c.Current_best_ips, expected_current_best_ips) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_ips, expected_current_best_ips)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(expected_time_of_last_evaluation) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\nexpected time\n%v", c.Time_of_last_evaluation, expected_time_of_last_evaluation)
	}
}

func TestFindBestHostsNoValidHostMinino(t *testing.T) {

	c := getTestCluster("testbad.cern.ch")

	c.Parameters.Metric = "minino"

	bad_hosts_to_check := getBadHostsToCheck(c)

	expected_current_best_ips := []net.IP{}
	isDNSUpdateValid, err := c.FindBestHosts(bad_hosts_to_check)
	if err != nil {
		t.Errorf("Error while finding best hosts. error:%v", err)
	}
	if !isDNSUpdateValid {
		t.Errorf("e.Find_best_hosts: returned false, expected true")
	}
	if !reflect.DeepEqual(c.Current_best_ips, expected_current_best_ips) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_ips, expected_current_best_ips)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(time.Now()) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\ncurrent time\n%v", c.Time_of_last_evaluation, time.Now())
	}
}

func TestFindBestHostsNoValidHostMinimum(t *testing.T) {

	c := getTestCluster("testbad.cern.ch")

	c.Parameters.Metric = "minimum"

	bad_hosts_to_check := getBadHostsToCheck(c)

	not_expected_current_best_ips := []net.IP{}
	isDNSUpdateValid, err := c.FindBestHosts(bad_hosts_to_check)
	if err != nil {
		t.Errorf("Error while finding best hosts. error:%v", err)
	}
	if !isDNSUpdateValid {
		t.Errorf("e.Find_best_hosts: returned false, expected true")
	}
	if reflect.DeepEqual(c.Current_best_ips, not_expected_current_best_ips) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nwhich is not expected", c.Current_best_ips)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(time.Now()) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\ncurrent time\n%v", c.Time_of_last_evaluation, time.Now())
	}
}
