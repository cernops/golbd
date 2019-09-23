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
	//Empty slice for comparisson purposes
	ipsEmpty := []string{}
	//Definition of expected hosts IP for aiermis, valid in the time when the test was written
	ExpectedIPAiermis := []string{"188.184.104.111", "2001:1458:d00:2d::100:58"}

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
		"testme007.cern.ch": {ipsEmpty, nil},
		"testme007":         {ipsEmpty, nil},
		"kkouros.cern.ch":   {ipsEmpty, nil},
		"aiermis.cern.ch":   {ExpectedIPAiermis, nil},
	}
	//receiving the output for every alias and storing the results into a map
	received := make(map[string][]interface{})
	iprecString := []string{}
	for _, c := range Clusters {
		iprec, err := c.Get_state_dns(dnsManager)
		for _, ip := range iprec {
			iprecString = append(iprecString, ip.String())
		}
		//Casting to string. The DeepEqual of  IP is a bit  tricky, since it can
		received[c.Cluster_name] = []interface{}{iprecString, err}
	}
	//DeepEqual comparison between the map with expected values and the one with the outputs
	for _, c := range Clusters {
		if !reflect.DeepEqual(received[c.Cluster_name], expected[c.Cluster_name]) {
			t.Errorf("\ngot\n%T and %v\nexpected\n%T and %v", received[c.Cluster_name][0], received[c.Cluster_name][0], expected[c.Cluster_name], expected[c.Cluster_name])
		}
	}

}
