package main

import (
	"os"
	"reflect"
	"testing"

	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbconfig"
	"lb-experts/golbd/logger"
)

func TestLoadConfig(t *testing.T) {

	lg, _ := logger.NewLoggerFactory("sample.log")
	lg.EnableWriteToSTd()

	configFromFile := lbconfig.NewLoadBalancerConfig("testloadconfig", lg)
	configExisting, _, err := configFromFile.Load()
	if err != nil {
		t.Fail()
		t.Errorf("loadConfig Error: %v", err.Error())
	}
	expConfig := lbconfig.NewLoadBalancerConfig("", lg)
	expConfig.SetMasterHost("lbdxyz.cern.ch")
	expConfig.SetHeartBeatFileName("heartbeat")
	expConfig.SetHeartBeatDirPath("/work/go/src/github.com/cernops/golbd")
	expConfig.SetTSIGKeyPrefix("abcd-")
	expConfig.SetTSIGInternalKey("xxx123==")
	expConfig.SetTSIGExternalKey("yyy123==")
	expConfig.SetDNSManager("137.138.28.176")
	expConfig.SetSNMPPassword("zzz123")
	expConfig.SetClusters(map[string][]string{
		"aiermis.cern.ch":     {"ermis19.cern.ch", "ermis20.cern.ch"},
		"uermis.cern.ch":      {"ermis21.cern.ch", "ermis22.cern.ch"},
		"permis.cern.ch":      {"ermis21.sub.cern.ch", "ermis22.test.cern.ch", "ermis42.cern.ch"},
		"ermis.test.cern.ch":  {"ermis23.cern.ch", "ermis24.cern.ch"},
		"ermis2.test.cern.ch": {"ermis23.toto.cern.ch", "ermis24.cern.ch", "ermis25.sub.cern.ch"},
	})
	expConfig.SetParameters(map[string]lbcluster.Params{
		"aiermis.cern.ch":     {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 60},
		"uermis.cern.ch":      {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
		"permis.cern.ch":      {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
		"ermis.test.cern.ch":  {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
		"ermis2.test.cern.ch": {Behaviour: "mindless", Best_hosts: 1, External: false, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long", Ttl: 222},
	})

	if !reflect.DeepEqual(configExisting, &expConfig) {
		t.Errorf("loadConfig: got\n %v expected\n %v", configExisting, &expConfig)
	}
	os.Remove("sample.log")
}

//func TestWatchConfigFileChanges(t *testing.T) {
//	lg := lbcluster.Log{Stdout: true, Debugflag: false}
//	var wg *sync.WaitGroup
//	var controlChan = make(chan bool)
//	defer close(controlChan)
//	config:=lbconfig2.NewLoadBalancerConfig("testloadconfig", &lg)
//	fileChangeSignal := config.WatchFileChange(controlChan, wg)
//	for filChangeData := range fileChangeSignal {
//
//	}
//}
