package main_test

import (
	"lb-experts/golbd/lbhost"
	"os"
	"reflect"
	"testing"
)

//Function TestGetLoadHosts tests the function GetLoadForAlias
func TestGetLoadHosts(t *testing.T) {

	hosts := []lbhost.Host{
		getHost("lxplus132.cern.ch", 7, ""),
		getHost("lxplus132.cern.ch", 0, "blabla.cern.ch=179,blablabla2.cern.ch=4"),
		getHost("toto132.lxplus.cern.ch", 42, ""),
		getHost("toto132.lxplus.cern.ch", 0, "blabla.subdo.cern.ch=179,blablabla2.subdo.cern.ch=4"),
	}

	expectedhost0 := hosts[0].GetHostTransportPayloads()[0].Response_int
	//expectedhost1 := hosts[1].HostTransports[0].Response_int
	expectedhost2 := hosts[2].GetHostTransportPayloads()[0].Response_int
	//expectedhost3 := hosts[3].HostTransports[0].Response_int

	if !reflect.DeepEqual(hosts[0].GetLoadForAlias(hosts[0].GetClusterConfig().Cluster_name), expectedhost0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[0].GetLoadForAlias(hosts[0].GetClusterConfig().Cluster_name), expectedhost0)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias(hosts[1].GetClusterConfig().Cluster_name), 0) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias(hosts[1].GetClusterConfig().Cluster_name), 0)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias("blabla.cern.ch"), 179) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias("blabla.cern.ch"), 179)
	}
	if !reflect.DeepEqual(hosts[1].GetLoadForAlias("blablabla2.cern.ch"), 4) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[1].GetLoadForAlias("blablabla2.cern.ch"), 4)
	}
	if !reflect.DeepEqual(hosts[2].GetLoadForAlias(hosts[2].GetClusterConfig().Cluster_name), expectedhost2) {
		t.Errorf(" got\n%v\nexpected\n%v", hosts[2].GetLoadForAlias(hosts[2].GetClusterConfig().Cluster_name), expectedhost2)
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
	err := os.Remove("sample.log")
	if err != nil {
		t.Fail()
		t.Errorf("error deleting file.error %v", err)
	}
}
