package main_test

import (
	"fmt"
	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
	"lb-experts/golbd/model"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
)

type mockHost struct {
}

func (m mockHost) GetHostTransportPayloads() []lbhost.LBHostTransportResult {
	panic("implement me")
}

func (m mockHost) SetName(name string) {
	panic("implement me")
}

func (m mockHost) SetTransportPayload(transportPayloadList []lbhost.LBHostTransportResult) {
	panic("implement me")
}

func (m mockHost) GetName() string {
	panic("implement me")
}

func (m mockHost) SNMPDiscovery() {
	panic("implement me")
}

func (m mockHost) GetClusterConfig() *model.CluserConfig {
	panic("implement me")
}

func (m mockHost) GetLoadForAlias(clusterName string) int {
	return 0
}

func (m mockHost) GetWorkingIPs() ([]net.IP, error) {
	return []net.IP{}, fmt.Errorf("sample error")
}

func (m mockHost) GetAllIPs() ([]net.IP, error) {
	panic("implement me")
}

func (m mockHost) GetIps() ([]net.IP, error) {
	time.Sleep(5 * time.Second) // simulating a network request
	return []net.IP{}, nil
}

func NewMockHost() lbhost.Host {
	return &mockHost{}
}

func compareIPs(t *testing.T, source, target []net.IP) {

	found := map[string]bool{}

	for _, b := range source {
		found[b.String()] = true
	}
	// Make sure that all the ips in the target are in source
	for _, value := range target {
		if _, ok := found[value.String()]; !ok {
			t.Errorf("The ip %v is not in the source list %v", value, source)
		}
		delete(found, value.String())
	}
	//If there are any elements left, fail
	if len(found) > 0 {
		t.Errorf("The ip(s) %v are not in the expected list %v", found, target)
	}
}
func compareHosts(t *testing.T, source, target map[string]lbcluster.Node) {
	for key, value := range source {
		if value.Load != target[key].Load {
			t.Errorf("Error comparing the list of hosts:\n The host %v is different:\n %v\n and\n %v\n", key, value, target[key])
		}
		compareIPs(t, value.IPs, target[key].IPs)
	}
	for key, _ := range target {
		if _, ok := source[key]; !ok {
			t.Errorf("Error comparing the list of hosts:\n The source doesn not have host %v (%v)\n", key, target[key])
		}
	}

}
func TestEvaluateHosts(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	hostsToCheck := getHostsToCheck(c)

	expectedHostMetricTable := map[string]lbcluster.Node{
		"lxplus130.cern.ch":               lbcluster.Node{Load: 27, IPs: []net.IP{net.ParseIP("188.184.108.100")}},
		"lxplus133.subdo.cern.ch":         lbcluster.Node{Load: 27, IPs: []net.IP{net.ParseIP("188.184.108.101")}},
		"lxplus132.cern.ch":               lbcluster.Node{Load: 2, IPs: []net.IP{net.ParseIP("2001:1458:d00:2c::100:a6"), net.ParseIP("188.184.108.98")}},
		"lxplus041.cern.ch":               lbcluster.Node{Load: 3, IPs: []net.IP{net.ParseIP("2001:1458:d00:32::100:51"), net.ParseIP("188.184.116.81")}},
		"monit-kafkax-17be060b0d.cern.ch": lbcluster.Node{Load: 816, IPs: []net.IP{net.ParseIP("188.184.108.100")}},
	}

	expectedCurrentBestIPs := c.Current_best_ips
	expectedTimeOfLastEvaluation := c.Time_of_last_evaluation

	c.EvaluateHosts(hostsToCheck)

	compareHosts(t, c.Host_metric_table, expectedHostMetricTable)

	compareIPs(t, c.Current_best_ips, expectedCurrentBestIPs)

	if !reflect.DeepEqual(c.Time_of_last_evaluation, expectedTimeOfLastEvaluation) {
		t.Errorf("e.evaluate_hosts: c.Time_of_last_evaluation: got\n%v\nexpected\n%v", c.Time_of_last_evaluation, expectedTimeOfLastEvaluation)
	}
}

func TestEvaluateHostsConcurrency(t *testing.T) {
	mockHostMap := make(map[string]lbhost.Host)
	mockHostMap["sampleHost"] = NewMockHost()
	logger, _ := logger.NewLoggerFactory("sample.log")
	cluster := lbcluster.LBCluster{Slog: logger}
	cluster.Host_metric_table = map[string]lbcluster.Node{"sampleHost": {HostName: "sampleHost"}}
	startTime := time.Now()
	cluster.EvaluateHosts(mockHostMap)
	endTime := time.Now()
	if endTime.Sub(startTime) > 6*time.Second {
		t.Fail()
		t.Errorf("concurrent job not running properly. expDuration:%v, actualDuration:%v", 6, endTime.Sub(startTime))
	}
	os.Remove("sample.log")
}
