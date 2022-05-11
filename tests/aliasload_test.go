package main_test

import (
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

//Function TestGetLoadHosts tests the function GetLoadForAlias
func TestGetLoadHosts(t *testing.T) {

	hosts := []lbhost.LBHost{
		getHost("lxplus132.cern.ch", 7, ""),
		getHost("lxplus132.cern.ch", 0, "blabla.cern.ch=179,blablabla2.cern.ch=4"),
		getHost("toto132.lxplus.cern.ch", 42, ""),
		getHost("toto132.lxplus.cern.ch", 0, "blabla.subdo.cern.ch=179,blablabla2.subdo.cern.ch=4"),
	}

	expectedhost0 := hosts[0].Host_transports[0].Response_int
	//expectedhost1 := hosts[1].HostTransports[0].Response_int
	expectedhost2 := hosts[2].Host_transports[0].Response_int
	//expectedhost3 := hosts[3].HostTransports[0].Response_int

	if !reflect.DeepEqual(hosts[0].Get_load_for_alias(hosts[0].Cluster_name), expectedhost0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[0].Get_load_for_alias(hosts[0].Cluster_name), expectedhost0)
	}
	if !reflect.DeepEqual(hosts[1].Get_load_for_alias(hosts[1].Cluster_name), 0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].Get_load_for_alias(hosts[1].Cluster_name), 0)
	}
	if !reflect.DeepEqual(hosts[1].Get_load_for_alias("blabla.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].Get_load_for_alias("blabla.cern.ch"), 179)
	}
	if !reflect.DeepEqual(hosts[1].Get_load_for_alias("blablabla2.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].Get_load_for_alias("blablabla2.cern.ch"), 4)
	}
	if !reflect.DeepEqual(hosts[2].Get_load_for_alias(hosts[2].Cluster_name), expectedhost2) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[2].Get_load_for_alias(hosts[2].Cluster_name), expectedhost2)
	}
	if !reflect.DeepEqual(hosts[2].Get_load_for_alias("toto.subdo.cern.ch"), expectedhost2) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[2].Get_load_for_alias("toto.subdo.cern.ch"), expectedhost2)
	}
	if !reflect.DeepEqual(hosts[3].Get_load_for_alias("blabla.subdo.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[3].Get_load_for_alias("blabla.subdo.cern.ch"), 179)
	}
	if !reflect.DeepEqual(hosts[3].Get_load_for_alias("blablabla2.subdo.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[3].Get_load_for_alias("blablabla2.subdo.cern.ch"), 4)
	}
}
