package main_test

import (
	"github.com/reguero/go-snmplib"
	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
	"net"
	"os"
	"testing"
	"time"
)

type mockSNMPAgent struct {
}

func (m mockSNMPAgent) Close() error {
	return nil
}

func (m mockSNMPAgent) Discover() error {
	time.Sleep(1 * time.Second)
	return nil
}

func (m mockSNMPAgent) GetV3(oid snmplib.Oid) (interface{}, error) {
	return 200, nil
}

func NewMockSNMPAgent() lbhost.DiscoveryAgent {
	return &mockSNMPAgent{}
}

func TestSNMPDiscoveryForConcurrency(t *testing.T) {
	lg, _ := logger.NewLoggerFactory("sample.log")
	lg.EnableWriteToSTd()
	host := lbhost.LBHost{Logger: lg, SnmpAgent: NewMockSNMPAgent()}
	host.HostTransports = []lbhost.LBHostTransportResult{
		{IP: net.ParseIP("1.1.1.1"), Transport: "udp"},
		{IP: net.ParseIP("1.1.1.2"), Transport: "udp"},
		{IP: net.ParseIP("1.1.1.3"), Transport: "udp"},
	}
	startTime := time.Now()
	host.SNMPDiscovery()
	endTime := time.Now()
	if endTime.Sub(startTime) > 2*time.Second {
		t.Fail()
		t.Errorf("execution took more time than expected. expectedTime: %v, actualTime:%v", 1, endTime.Sub(startTime))
	}
	os.Remove("sample.log")
}
