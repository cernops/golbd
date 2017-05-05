package main

import (
	"github.com/cernops/golbd/lbcluster"
	//"sync"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}
	expected := Config{Master: "lbdxyz.cern.ch",
		HeartbeatFile: "heartbeat",
		HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
		//HeartbeatMu:     sync.Mutex{0, 0},
		TsigKeyPrefix:   "abcd-",
		TsigInternalKey: "xxx123==",
		TsigExternalKey: "yyy123==",
		SnmpPassword:    "zzz123",
		DnsManager:      "111.111.0.111",
		Clusters: map[string][]string{"test01.cern.ch": []string{"lxplus142.cern.ch", "lxplus177.cern.ch"},
			"test02.cern.ch": []string{"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus025.cern.ch"}},
		Parameters: map[string]lbcluster.Params{"test01.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			"test02.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"}}}

	config, e := loadConfig("./loadConfigtst.conf", lg)
	if e != nil {
		t.Errorf("loadConfig Error: %v", e.Error())
	} else {
		if !reflect.DeepEqual(config, expected) {
			t.Errorf("loadConfig: got %v expected %v", config, expected)
		}
	}

}
