package main_test

import (
	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
	"lb-experts/golbd/model"
	"net"
	"os"
	"reflect"
	"testing"
)

func TestGetListHostsOne(t *testing.T) {
	c := getTestCluster("test01.cern.ch")
	host1 := lbhost.NewLBHost(c.ClusterConfig, c.Slog)
	host1.SetName("lxplus041.cern.ch")
	host2 := lbhost.NewLBHost(c.ClusterConfig, c.Slog)
	host2.SetName("monit-kafkax-17be060b0d.cern.ch")
	host3 := lbhost.NewLBHost(c.ClusterConfig, c.Slog)
	host3.SetName("lxplus132.cern.ch")
	host4 := lbhost.NewLBHost(c.ClusterConfig, c.Slog)
	host4.SetName("lxplus041.cern.ch")
	host5 := lbhost.NewLBHost(c.ClusterConfig, c.Slog)
	host5.SetName("lxplus041.cern.ch")
	expected := map[string]lbhost.Host{
		"lxplus041.cern.ch":               host1,
		"monit-kafkax-17be060b0d.cern.ch": host2,
		"lxplus132.cern.ch":               host3,
		"lxplus133.subdo.cern.ch":         host4,
		"lxplus130.cern.ch":               host5,
	}

	hostsToCheck := c.GetHostList()
	if len(hostsToCheck) != len(expected) {
		t.Errorf("length mismatch. expected :%v, actual:%v", len(expected), len(hostsToCheck))
	}
	for hostName, actualHost := range hostsToCheck {
		expHost := expected[hostName]
		if !reflect.DeepEqual(expHost.GetClusterConfig(), actualHost.GetClusterConfig()) {
			t.Errorf("mismatch in cluster config. expected:%v,actual:%v", expHost.GetClusterConfig(), actualHost.GetClusterConfig())
		}
	}
}

func TestGetListHostsTwo(t *testing.T) {
	logger, _ := logger.NewLoggerFactory("sample.log")
	logger.EnableWriteToSTd()

	clusters := []lbcluster.LBCluster{
		{ClusterConfig: model.CluserConfig{
			Cluster_name:           "test01.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
		}, Host_metric_table: map[string]lbcluster.Node{"lxplus142.cern.ch": lbcluster.Node{}, "lxplus177.cern.ch": lbcluster.Node{}},
			Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_ips: []net.IP{},

			Previous_best_ips_dns: []net.IP{},
			Slog:                  logger,
			Current_index:         0},
		lbcluster.LBCluster{ClusterConfig: model.CluserConfig{
			Cluster_name:           "test02.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
		},
			Host_metric_table: map[string]lbcluster.Node{"lxplus013.cern.ch": lbcluster.Node{}, "lxplus177.cern.ch": lbcluster.Node{}, "lxplus025.cern.ch": lbcluster.Node{}},
			Parameters:        lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_ips:      []net.IP{},
			Previous_best_ips_dns: []net.IP{},
			Slog:                  logger,
			Current_index:         0}}

	host1 := lbhost.NewLBHost(model.CluserConfig{
		Cluster_name:           "test01.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
	}, logger)
	host1.SetName("lxplus142.cern.ch")
	host2 := lbhost.NewLBHost(model.CluserConfig{
		Cluster_name:           "test01.cern.ch,test02.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
	}, logger)
	host2.SetName("lxplus177.cern.ch")
	host3 := lbhost.NewLBHost(model.CluserConfig{
		Cluster_name:           "test02.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
	}, logger)
	host3.SetName("lxplus013.cern.ch")
	host4 := lbhost.NewLBHost(model.CluserConfig{
		Cluster_name:           "test02.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
	}, logger)
	host4.SetName("lxplus025.cern.ch")
	expected := map[string]lbhost.Host{
		"lxplus142.cern.ch": host1,
		"lxplus177.cern.ch": host2,
		"lxplus013.cern.ch": host3,
		"lxplus025.cern.ch": host4,
	}

	var hostsToCheck map[string]lbhost.Host
	for _, c := range clusters {
		hostsToCheck = c.GetHostList()
	}
	if !reflect.DeepEqual(hostsToCheck, expected) {
		t.Errorf("e.GetHostList: got\n%v\nexpected\n%v", hostsToCheck, expected)
	}
	os.Remove("sample.log")
}
