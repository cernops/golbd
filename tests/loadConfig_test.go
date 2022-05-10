package main

import (
	"os"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbconfig"
)

func TestLoadConfig(t *testing.T) {
	lg := lbcluster.Log{Stdout: true, Debugflag: false}

	testFiles := []string{"testloadconfig.yaml", "testloadconfig"}

	//open files
	for _, testFile := range testFiles {
		loadconfig, err := os.Open(testFile)
		if err != nil {
			panic(err)
		}
		defer loadconfig.Close()

		// The expected output
		expected :=
			lbconfig.Config{
				Master:        "lbdxyz.cern.ch",
				HeartbeatFile: "heartbeat",
				HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
				//HeartbeatMu:     sync.Mutex{0, 0},
				TsigKeyPrefix:   "abcd-",
				TsigInternalKey: "xxx123==",
				TsigExternalKey: "yyy123==",
				SnmpPassword:    "zzz123",
				DNSManager:      "137.138.28.176:53",
				ConfigFile:      testFile,
				Clusters: map[string][]string{
					"aiermis.cern.ch":     {"ermis19.cern.ch", "ermis20.cern.ch"},
					"uermis.cern.ch":      {"ermis21.cern.ch", "ermis22.cern.ch"},
					"permis.cern.ch":      {"ermis21.sub.cern.ch", "ermis22.test.cern.ch", "ermis42.cern.ch"},
					"ermis.test.cern.ch":  {"ermis23.cern.ch", "ermis24.cern.ch"},
					"ermis2.test.cern.ch": {"ermis23.toto.cern.ch", "ermis24.cern.ch", "ermis25.sub.cern.ch"}},
				Parameters: map[string]lbcluster.Params{
					"aiermis.cern.ch":     {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 60},
					"uermis.cern.ch":      {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
					"permis.cern.ch":      {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
					"ermis.test.cern.ch":  {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
					"ermis2.test.cern.ch": {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222}}}

		//retrieving the actual output
		configExisting, _, e := lbconfig.LoadConfig(loadconfig.Name(), &lg)

		if e != nil {
			t.Errorf("loadConfig Error: %v", e.Error())
		} else {
			if !reflect.DeepEqual(configExisting, &expected) {
				t.Errorf("loadConfig: got\n %+v \nexpected\n %+v", configExisting, &expected)
			}
		}
	}
}
