package main

import (
	"os"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
)

func TestLoadConfig(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

	//open file
	loadconfig, err := os.Open("testloadconfig")
	if err != nil {
		panic(err)
	}
	defer loadconfig.Close()

	// The expected output
	expected :=
		Config{
			Master:        "lbdxyz.cern.ch",
			HeartbeatFile: "heartbeat",
			HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
			//HeartbeatMu:     sync.Mutex{0, 0},
			TsigKeyPrefix:   "abcd-",
			TsigInternalKey: "xxx123==",
			TsigExternalKey: "yyy123==",
			SnmpPassword:    "zzz123",
			DNSManager:      "137.138.28.176",
			Clusters: map[string][]string{
				"aiermis.cern.ch": {"ermis19.cern.ch", "ermis20.cern.ch"},
				"uermis.cern.ch":  {"ermis21.cern.ch", "ermis22.cern.ch"}},
			Parameters: map[string]lbcluster.Params{
				"aiermis.cern.ch": {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 60},
				"uermis.cern.ch":  {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222}}}

	//retrieving the actual output
	configExisting, _, e := loadConfig(loadconfig.Name(), &lg)

	if e != nil {
		t.Errorf("loadConfig Error: %v", e.Error())
	} else {
		if !reflect.DeepEqual(configExisting, &expected) {
			t.Errorf("loadConfig: got %v expected %v", configExisting, &expected)
		}

	}

}
