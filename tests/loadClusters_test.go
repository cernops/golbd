package main_test

import (
	"lb-experts/golbd/lbcluster"
	"lb-experts/golbd/lbconfig"
	"lb-experts/golbd/lbhost"
	"lb-experts/golbd/logger"
	"lb-experts/golbd/model"
	"net"
	"os"
	"reflect"
	"testing"
)

func getTestCluster(name string) lbcluster.LBCluster {
	lg, _ := logger.NewLoggerFactory("sample.log")

	return lbcluster.LBCluster{
		ClusterConfig: model.CluserConfig{
			Cluster_name:           name,
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
		},
		Host_metric_table: map[string]lbcluster.Node{
			"lxplus132.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus041.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus130.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus133.subdo.cern.ch":         lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"monit-kafkax-17be060b0d.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}}},
		Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_ips:      []net.IP{},
		Previous_best_ips_dns: []net.IP{},
		Slog:                  lg,
		Current_index:         0}
}

func getSecondTestCluster() lbcluster.LBCluster {
	lg, _ := logger.NewLoggerFactory("sample.log")

	return lbcluster.LBCluster{
		ClusterConfig: model.CluserConfig{
			Cluster_name:           "test02.test.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
		},
		Host_metric_table: map[string]lbcluster.Node{
			"lxplus013.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus038.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus039.test.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus025.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}}},
		Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_ips:      []net.IP{},
		Previous_best_ips_dns: []net.IP{},
		Slog:                  lg,
		Current_index:         0}
}
func getHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.Host {
	lg, _ := logger.NewLoggerFactory("sample.log")
	host1 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host1.SetName("lxplus132.cern.ch")
	host1.SetTransportPayload([]lbhost.LBHostTransportResult{
		lbhost.LBHostTransportResult{Transport: "udp6", Response_int: 2, Response_string: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), Response_error: ""},
		lbhost.LBHostTransportResult{Transport: "udp", Response_int: 2, Response_string: "", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
	})
	host2 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host2.SetName("lxplus041.cern.ch")
	host2.SetTransportPayload([]lbhost.LBHostTransportResult{
		lbhost.LBHostTransportResult{Transport: "udp6", Response_int: 3, Response_string: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), Response_error: ""},
		lbhost.LBHostTransportResult{Transport: "udp", Response_int: 3, Response_string: "", IP: net.ParseIP("188.184.116.81"), Response_error: ""},
	})
	host3 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host3.SetName("lxplus130.cern.ch")
	host3.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: 27, Response_string: "", IP: net.ParseIP("188.184.108.100"), Response_error: "",
	}})
	host4 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host4.SetName("lxplus133.subdo.cern.ch")
	host4.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: 27, Response_string: "", IP: net.ParseIP("188.184.108.101"), Response_error: "",
	}})
	host5 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host5.SetName("monit-kafkax-17be060b0d.cern.ch")
	host5.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: 100000, Response_string: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
	)
	hostsToCheck := map[string]lbhost.Host{
		"lxplus132.cern.ch":               host1,
		"lxplus041.cern.ch":               host2,
		"lxplus130.cern.ch":               host3,
		"lxplus133.subdo.cern.ch":         host4,
		"monit-kafkax-17be060b0d.cern.ch": host5,
	}

	return hostsToCheck
}
func getBadHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.Host {
	lg, _ := logger.NewLoggerFactory("sample.log")
	host1 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host1.SetName("lxplus132.cern.ch")
	host1.SetTransportPayload([]lbhost.LBHostTransportResult{
		lbhost.LBHostTransportResult{Transport: "udp6", Response_int: -2, Response_string: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), Response_error: ""},
		lbhost.LBHostTransportResult{Transport: "udp", Response_int: -2, Response_string: "", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
	})
	host2 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host2.SetName("lxplus041.cern.ch")
	host2.SetTransportPayload([]lbhost.LBHostTransportResult{
		lbhost.LBHostTransportResult{Transport: "udp6", Response_int: -3, Response_string: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), Response_error: ""},
		lbhost.LBHostTransportResult{Transport: "udp", Response_int: -3, Response_string: "", IP: net.ParseIP("188.184.116.81"), Response_error: ""},
	})
	host3 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host3.SetName("lxplus130.cern.ch")
	host3.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: -27, Response_string: "", IP: net.ParseIP("188.184.108.100"), Response_error: "",
	}})
	host4 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host4.SetName("lxplus133.subdo.cern.ch")
	host4.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: -15, Response_string: "", IP: net.ParseIP("188.184.108.101"), Response_error: "",
	}})
	host5 := lbhost.NewLBHost(c.ClusterConfig, lg)
	host5.SetName("monit-kafkax-17be060b0d.cern.ch")
	host5.SetTransportPayload([]lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{
		Transport: "udp", Response_int: 100000, Response_string: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
	)
	badHostsToCheck := map[string]lbhost.Host{
		"lxplus132.cern.ch":               host1,
		"lxplus041.cern.ch":               host2,
		"lxplus130.cern.ch":               host3,
		"lxplus133.subdo.cern.ch":         host4,
		"monit-kafkax-17be060b0d.cern.ch": host5,
	}

	return badHostsToCheck
}
func getHost(hostname string, responseInt int, responseString string) lbhost.Host {
	lg, _ := logger.NewLoggerFactory("sample.log")
	clusterConfig := model.CluserConfig{
		Cluster_name:           "test01.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "XXXX",
	}
	host1 := lbhost.NewLBHost(clusterConfig, lg)
	host1.SetName(hostname)
	host1.SetTransportPayload([]lbhost.LBHostTransportResult{
		lbhost.LBHostTransportResult{Transport: "udp", Response_int: responseInt, Response_string: responseString, IP: net.ParseIP("188.184.108.98"), Response_error: ""}},
	)
	return host1

}
func TestLoadClusters(t *testing.T) {
	lg, _ := logger.NewLoggerFactory("sample.log")
	lg.EnableWriteToSTd()

	config := lbconfig.NewLoadBalancerConfig("", lg)
	config.SetMasterHost("lbdxyz.cern.ch")
	config.SetHeartBeatFileName("heartbeat")
	config.SetHeartBeatDirPath("/work/go/src/github.com/cernops/golbd")
	config.SetTSIGKeyPrefix("abcd-")
	config.SetTSIGInternalKey("xxx123==")
	config.SetTSIGExternalKey("yyy123==")
	config.SetDNSManager("111.111.0.111")
	config.SetSNMPPassword("zzz123")
	config.SetClusters(map[string][]string{
		"test01.cern.ch":      {"lxplus132.cern.ch", "lxplus041.cern.ch", "lxplus130.cern.ch", "lxplus133.subdo.cern.ch", "monit-kafkax-17be060b0d.cern.ch"},
		"test02.test.cern.ch": {"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus039.test.cern.ch", "lxplus025.cern.ch"},
	})
	config.SetParameters(map[string]lbcluster.Params{
		"test01.cern.ch":      lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		"test02.test.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
	})

	expected := []lbcluster.LBCluster{getTestCluster("test01.cern.ch"),
		getSecondTestCluster()}

	lbclusters, _ := config.Load()
	// reflect.DeepEqual(lbclusters, expected) occassionally fails as the array order is not always the same
	// so comparing element par element
	i := 0
	for _, e := range expected {
		for _, c := range lbclusters {
			if c.ClusterConfig.Cluster_name == e.ClusterConfig.Cluster_name {
				if !reflect.DeepEqual(c, e) {
					t.Errorf("loadClusters: got\n%v\nexpected\n%v", lbclusters, expected)
				} else {
					i = i + 1
				}
				continue
			}
		}
	}
	if (i != len(expected)) || (i != len(lbclusters)) {
		t.Errorf("loadClusters: wrong number of clusters, got\n%v\nexpected\n%v (and %v", len(lbclusters), len(expected), i)

	}
	err := os.Remove("sample.log")
	if err != nil {
		t.Fail()
		t.Errorf("error deleting file.error %v", err)
	}
}
