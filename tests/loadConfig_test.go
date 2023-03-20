package main

import (
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbconfig"
	"lb-experts/golbd/logger"
)

func TestLoadConfig(t *testing.T) {

	testFiles := []string{"testloadconfig.yaml", "testloadconfig"}
	lg, _ := logger.NewLoggerFactory("sample.log")
	lg.EnableWriteToSTd()
	for _, testFile := range testFiles {
		configFromFile := lbconfig.NewLoadBalancerConfig(testFile, lg)
		_, err := configFromFile.Load()
		if err != nil {
			t.Fail()
			t.Errorf("loadConfig Error: %v", err.Error())
		}
		expConfig := lbconfig.NewLoadBalancerConfig(testFile, lg)
		expConfig.SetMasterHost("lbdxyz.cern.ch")
		expConfig.SetHeartBeatFileName("heartbeat")
		expConfig.SetHeartBeatDirPath("/work/go/src/github.com/cernops/golbd")
		expConfig.SetTSIGKeyPrefix("abcd-")
		expConfig.SetTSIGInternalKey("xxx123==")
		expConfig.SetTSIGExternalKey("yyy123==")
		expConfig.SetDNSManager("137.138.28.176:53")
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

		if !reflect.DeepEqual(configFromFile, expConfig) {
			t.Errorf("loadConfig: got\n %v expected\n %v", configFromFile, expConfig)
		}

	}
	os.Remove("sample.log")

}
func TestWatchConfigFileChanges(t *testing.T) {
	lg, _ := logger.NewLoggerFactory("sample.log")
	lg.EnableWriteToSTd()
	var wg sync.WaitGroup
	var controlChan = make(chan bool)
	var changeCounter int
	sampleConfigFileName := "sampleConfig"
	dataSet := []string{
		"data 1",
		"data 12",
		"data 123",
	}

	err := createTestConfigFile(sampleConfigFileName)
	if err != nil {
		t.Fail()
		t.Errorf("error while creating test config file. name: %s", sampleConfigFileName)
	}
	go func() {
		defer close(controlChan)
		for _, dataToWrite := range dataSet {
			time.Sleep(1 * time.Second)
			err = writeDataToFile(sampleConfigFileName, dataToWrite)
			if err != nil {
				t.Fail()
				t.Errorf("error while writting to test config file. filename: %s, data:%s", sampleConfigFileName, dataToWrite)
			}
		}
	}()
	config := lbconfig.NewLoadBalancerConfig(sampleConfigFileName, lg)
	fileChangeSignal := config.WatchFileChange(controlChan, wg)
	for fileChangeData := range fileChangeSignal {
		changeCounter += 1
		t.Log("file change signal", fileChangeData)
	}
	if changeCounter == 0 {
		t.Fail()
		t.Error("file changes not observed")
	}
	deleteFile("sample.log")
	deleteFile(sampleConfigFileName)
}

func createTestConfigFile(fileName string) error {
	_, err := os.Create(fileName)
	if err != nil {
		return err
	}
	return nil
}

func writeDataToFile(fileName string, data string) error {
	fp, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = fp.WriteString(data)
	fp.Close()
	return err
}

func deleteFile(fileName string) {
	os.Remove(fileName)
}
