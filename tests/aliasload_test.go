package main_test

import (
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

// Function TestGetLoadHosts tests the function GetLoadForAlias
func TestGetLoadHosts(t *testing.T) {

	hosts := []lbhost.LBHost{
		getHost("lxplus132.cern.ch", 7, ""),
		getHost("lxplus132.cern.ch", 0, "blabla.cern.ch=179,blablabla2.cern.ch=4"),
		getHost("toto132.lxplus.cern.ch", 42, ""),
		getHost("toto132.lxplus.cern.ch", 0, "blabla.subdo.cern.ch=179,blablabla2.subdo.cern.ch=4"),
	}

	expectedhost0 := hosts[0].HostTransports[0].ResponseInt
	//expectedhost1 := hosts[1].Host_transports[0].Response_int
	expectedhost2 := hosts[2].HostTransports[0].ResponseInt
	//expectedhost3 := hosts[3].Host_transports[0].Response_int

	if !reflect.DeepEqual(hosts[0].GetLoadForAlias(hosts[0].ClusterName), expectedhost0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[0].GetLoadForAlias(hosts[0].ClusterName), expectedhost0)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias(hosts[1].ClusterName), 0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias(hosts[1].ClusterName), 0)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias("blabla.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias("blabla.cern.ch"), 179)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias("blablabla2.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias("blablabla2.cern.ch"), 4)
	}
	if !reflect.DeepEqual(hosts[2].GetLoadForAlias(hosts[2].ClusterName), expectedhost2) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[2].GetLoadForAlias(hosts[2].ClusterName), expectedhost2)
	}
	if !reflect.DeepEqual(hosts[2].GetLoadForAlias("toto.subdo.cern.ch"), expectedhost2) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[2].GetLoadForAlias("toto.subdo.cern.ch"), expectedhost2)
	}
	if !reflect.DeepEqual(hosts[3].GetLoadForAlias("blabla.subdo.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[3].GetLoadForAlias("blabla.subdo.cern.ch"), 179)
	}
	if !reflect.DeepEqual(hosts[3].GetLoadForAlias("blablabla2.subdo.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[3].GetLoadForAlias("blablabla2.subdo.cern.ch"), 4)
	}
}
