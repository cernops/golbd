package main_test

import (
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
)

//TestGetStateDNS tests the function get_state_dns
func TestGetStateDNS(t *testing.T) {
	//DNS IP
	dnsManager := "137.138.16.5"

	Clusters := []lbcluster.LBCluster{
		//Non-existing clusters
		getTestCluster("testme007.cern.ch"),
		getTestCluster("testme007"),
		//Existing clusters
		getTestCluster("kkouros.cern.ch"),
		getTestCluster("aiermis.cern.ch"),
	}
	//Expected response for every alias ( slice of IP and error message if any)
	expected := map[string][]interface{}{
		"testme007.cern.ch": {[]string{}, nil},
		"testme007":         {[]string{}, nil},
		"kkouros.cern.ch":   {[]string{}, nil},
		"aiermis.cern.ch":   {[]string{}, nil},
	}
	//receiving the output for every alias and storing the results into a map
	received := make(map[string][]interface{})
	iprecString := []string{}
	for _, c := range Clusters {
		err := c.GetStateDNS(dnsManager)
		iprec := c.Previous_best_ips_dns
		for _, ip := range iprec {
			iprecString = append(iprecString, ip.String())
		}
		//Casting to string. The DeepEqual of  IP is a bit  tricky, since it can
		received[c.Cluster_name] = []interface{}{iprecString, err}
	}
	//DeepEqual comparison between the map with expected values and the one with the outputs
	for _, c := range Clusters {
		if !reflect.DeepEqual(received[c.Cluster_name], expected[c.Cluster_name]) {
			t.Errorf("\ngot ips\n%T type and value %v\nexpected\n%T type and value %v", received[c.Cluster_name][0], received[c.Cluster_name][0], expected[c.Cluster_name][0], expected[c.Cluster_name][0])
			t.Errorf("\ngot error\n%T type and value %v\nexpected\n%T type and value %v", received[c.Cluster_name][1], received[c.Cluster_name][1], expected[c.Cluster_name][1], expected[c.Cluster_name][1])
		}
	}

}
