package main_test

import (
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

//Function TestGetLoadHosts tests the function Get_load_for_alias
func TestGetLoadHosts(t *testing.T) {

	hosts := []lbhost.LBHost{
		getHost("lxplus132.cern.ch", 7, ""),
		getHost("lxplus132.cern.ch", 0, "blabla.cern.ch=179,blablabla2.cern.ch=4"),
	}

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
