package lbhost

import (
	"net"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

//Function TestGetLoadHosts tests the function Get_load_for_alias
func TestGetLoadHosts(t *testing.T) {

	hosts := []lbhost.LBHost{
		{Cluster_name: "test01.cern.ch",
			Host_name: "lxplus132.cern.ch",
			Host_transports: []lbhost.LBHostTransportResult{
				lbhost.LBHostTransportResult{Transport: "udp", Response_int: 7, Response_string: "", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
			},
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			LogFile:                "",
			Debugflag:              false,
		},
		lbhost.LBHost{
			Cluster_name: "blabla.cern.ch",
			Host_name:    "lxplus013.cern.ch",
			Host_transports: []lbhost.LBHostTransportResult{
				lbhost.LBHostTransportResult{Transport: "udp", Response_int: 0, Response_string: "blabla.cern.ch=179,blablabla2.cern.ch=4", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
			},
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "XXXX",
			LogFile:                "",
			Debugflag:              false,
		}}

	expectedhost0 := hosts[0].Host_transports[0].Response_int
	expectedhost1 := hosts[1].Host_transports[0].Response_int

	if !reflect.DeepEqual(hosts[0].Get_load_for_alias(hosts[0].Cluster_name), expectedhost0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[0].Get_load_for_alias(hosts[0].Cluster_name), expectedhost0)
	}
	if !reflect.DeepEqual(hosts[1].Get_load_for_alias("blabla.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].Get_load_for_alias(hosts[1].Cluster_name), expectedhost1)
	}
	if !reflect.DeepEqual(hosts[1].Get_load_for_alias("blablabla2.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].Get_load_for_alias(hosts[1].Cluster_name), expectedhost1)
	}
}
