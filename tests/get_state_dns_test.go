package main_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
)

//TestGetStateDNS tests the function GetStateDNS
func TestGetStateDNS(t *testing.T) {
	// Create a local dns server
	server, err := setupDnsServer("5354")
	if err != nil {
		t.Errorf("Failed to setup DNS server for the test.")
	}
	defer server.Shutdown()

	//DNS IP
	dnsManager := "127.0.0.1:5354"

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
		"aiermis.cern.ch":   {[]string{"188.184.104.111", "2001:1458:d00:2d::100:58"}, nil},
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

//TestRefreshDNS tests the function RefreshDNS
func TestRefreshDNS(t *testing.T) {
	// Create a local dns server
	server, err := setupDnsServer("5353")
	if err != nil {
		t.Errorf("Failed to setup DNS server for the test.")
	}
	defer server.Shutdown()

	//DNS IP
	dnsManager := "127.0.0.1:5353"

	tests := []struct {
		cluster_name     string
		current_best_ips []net.IP
	}{
		{"aiermis.cern.ch", []net.IP{net.ParseIP("189.184.104.222"), net.ParseIP("3001:1458:d00:2d::100:59")}},
		{"testrefresh.cern.ch", []net.IP{net.ParseIP("2.3.4.5")}},
		{"nochange.cern.ch", []net.IP{net.ParseIP("1.1.1.1")}},
		{"notexists.cern.ch", []net.IP{net.ParseIP("2.2.2.2")}},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test_case_%d", i), func(t *testing.T) {
			lg := lbcluster.Log{SyslogWriter: nil, Stdout: false, Debugflag: false}
			cluster := lbcluster.LBCluster{
				Cluster_name:          tc.cluster_name,
				Current_best_ips:      tc.current_best_ips,
				Previous_best_ips_dns: []net.IP{},
				Slog:                  &lg,
			}

			cluster.RefreshDNS(dnsManager, "abcd-", "xxx123==", "yyy123==")
			cluster.GetStateDNS(dnsManager)

			got := cluster.Previous_best_ips_dns
			expected := tc.current_best_ips
			if !reflect.DeepEqual(expected, got) {
				t.Fatalf("test %d: expected: %v, got: %v", i+1, expected, got)
			}
		})
	}
}
