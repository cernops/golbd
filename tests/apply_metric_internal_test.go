package main_test

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"
)

func TestEvaluateMetric(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	c.EvaluateHosts(getHostsToCheck(c))

	expected_time_of_last_evaluation := c.Time_of_last_evaluation

	var myTests map[int][]net.IP
	myTests = make(map[int][]net.IP)

	//Getting the best two nodes
	myTests[2] = []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"),
		net.ParseIP("188.184.108.98"),
		net.ParseIP("2001:1458:d00:32::100:51"),
		net.ParseIP("188.184.116.81"),
	}
	// Only the ips of the best node
	myTests[1] = []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"),
		net.ParseIP("188.184.108.98")}

	//With -1, we should get all the nodes
	myTests[-1] = []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"),
		net.ParseIP("188.184.108.98"),
		net.ParseIP("2001:1458:d00:32::100:51"),
		net.ParseIP("188.184.116.81"),
		net.ParseIP("188.184.108.100"),
		net.ParseIP("188.184.108.101"),
	}

	for best, ips := range myTests {
		fmt.Printf("Checking if with %v best host it works", best)
		c.Parameters.Best_hosts = best
		c.ApplyMetric(getHostsToCheck(c))
		compareIPs(t, c.Current_best_ips, ips)
		//		if !reflect.DeepEqual(c.Current_best_ips, ips) {
		//			t.Errorf("e.apply_metric: Best:%v c.Current_best_ips: got\n %v\nexpected\n%v", best, c.Current_best_ips, ips)
		//}
	}
	if !reflect.DeepEqual(c.Time_of_last_evaluation, expected_time_of_last_evaluation) {
		t.Errorf("e.apply_metric: c.Time_of_last_evaluation: got\n%v\nexpected\n%v", c.Time_of_last_evaluation, expected_time_of_last_evaluation)
	}
	err := os.Remove("sample.log")
	if err != nil {
		t.Fail()
		t.Errorf("error deleting file.error %v", err)
	}
}
