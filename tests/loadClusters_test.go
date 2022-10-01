package main_test

import (
	"net"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbconfig"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

func getTestCluster(name string) lbcluster.LBCluster {
	lg := lbcluster.Log{SyslogWriter: nil, Stdout: true, Debugflag: false}
	return lbcluster.LBCluster{Cluster_name: name,
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
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
		Slog:                  &lg,
		Current_index:         0}
}

func getSecondTestCluster() lbcluster.LBCluster {
	lg := lbcluster.Log{SyslogWriter: nil, Stdout: true, Debugflag: false}
	return lbcluster.LBCluster{Cluster_name: "test02.test.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
		Host_metric_table: map[string]lbcluster.Node{
			"lxplus013.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus038.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus039.test.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus025.cern.ch":      lbcluster.Node{Load: 100000, IPs: []net.IP{}}},
		Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_ips:      []net.IP{},
		Previous_best_ips_dns: []net.IP{},
		Slog:                  &lg,
		Current_index:         0}
}
func getHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.LBHost {
	hostsToCheck := map[string]lbhost.LBHost{
		"lxplus132.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName: "lxplus132.cern.ch",
			HostTransports: []lbhost.TransportResult{
				lbhost.TransportResult{Transport: "udp6", ResponseInt: 2, ResponseString: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), ResponseError: ""},
				lbhost.TransportResult{Transport: "udp", ResponseInt: 2, ResponseString: "", IP: net.ParseIP("188.184.108.98"), ResponseError: ""},
			},
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus041.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName: "lxplus041.cern.ch",
			HostTransports: []lbhost.TransportResult{
				lbhost.TransportResult{Transport: "udp6", ResponseInt: 3, ResponseString: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), ResponseError: ""},
				lbhost.TransportResult{Transport: "udp", ResponseInt: 3, ResponseString: "", IP: net.ParseIP("188.184.116.81"), ResponseError: ""},
			},
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus130.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "lxplus130.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: 27, ResponseString: "", IP: net.ParseIP("188.184.108.100"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
		"lxplus133.subdo.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "lxplus130.subdo.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: 27, ResponseString: "", IP: net.ParseIP("188.184.108.101"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
		"monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "monit-kafkax-17be060b0d.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: 100000, ResponseString: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
	}

	return hostsToCheck
}
func getBadHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.LBHost {
	badHostsToCheck := map[string]lbhost.LBHost{
		"lxplus132.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName: "lxplus132.cern.ch",
			HostTransports: []lbhost.TransportResult{
				lbhost.TransportResult{Transport: "udp6", ResponseInt: -2, ResponseString: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), ResponseError: ""},
				lbhost.TransportResult{Transport: "udp", ResponseInt: -2, ResponseString: "", IP: net.ParseIP("188.184.108.98"), ResponseError: ""},
			},
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus041.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName: "lxplus041.cern.ch",
			HostTransports: []lbhost.TransportResult{
				lbhost.TransportResult{Transport: "udp6", ResponseInt: -3, ResponseString: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), ResponseError: ""},
				lbhost.TransportResult{Transport: "udp", ResponseInt: -3, ResponseString: "", IP: net.ParseIP("188.184.116.81"), ResponseError: ""},
			},
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus130.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "lxplus130.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: -27, ResponseString: "", IP: net.ParseIP("188.184.108.100"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
		"lxplus133.subdo.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "lxplus133.subdo.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: -15, ResponseString: "", IP: net.ParseIP("188.184.108.101"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
		"monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:       "monit-kafkax-17be060b0d.cern.ch",
			HostTransports: []lbhost.TransportResult{lbhost.TransportResult{Transport: "udp", ResponseInt: 100000, ResponseString: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), ResponseError: ""}},
			LBUsername:     c.Loadbalancing_username,
			LBPassword:     c.Loadbalancing_password,
			LogFile:        c.Slog.TofilePath,
			DebugFlag:      c.Slog.Debugflag,
		},
	}

	return badHostsToCheck
}
func getHost(hostname string, responseInt int, responseString string) lbhost.LBHost {

	return lbhost.LBHost{ClusterName: "test01.cern.ch",
		HostName: hostname,
		HostTransports: []lbhost.TransportResult{
			lbhost.TransportResult{Transport: "udp", ResponseInt: responseInt, ResponseString: responseString, IP: net.ParseIP("188.184.108.98"), ResponseError: ""}},
		LBUsername: "loadbalancing",
		LBPassword: "XXXX",
		LogFile:    "",
		DebugFlag:  false,
	}

}
func TestLoadClusters(t *testing.T) {
	lg := lbcluster.Log{SyslogWriter: nil, Stdout: true, Debugflag: false}

	config := lbconfig.Config{Master: "lbdxyz.cern.ch",
		HeartbeatFile: "heartbeat",
		HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
		//HeartbeatMu:     sync.Mutex{0, 0},
		TsigKeyPrefix:   "abcd-",
		TsigInternalKey: "xxx123==",
		TsigExternalKey: "yyy123==",
		SnmpPassword:    "zzz123",
		DNSManager:      "111.111.0.111:53",
		Clusters:        map[string][]string{"test01.cern.ch": {"lxplus132.cern.ch", "lxplus041.cern.ch", "lxplus130.cern.ch", "lxplus133.subdo.cern.ch", "monit-kafkax-17be060b0d.cern.ch"}, "test02.test.cern.ch": {"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus039.test.cern.ch", "lxplus025.cern.ch"}},
		Parameters: map[string]lbcluster.Params{"test01.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 2,
			External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			"test02.test.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"}}}
	expected := []lbcluster.LBCluster{getTestCluster("test01.cern.ch"),
		getSecondTestCluster()}

	lbclusters, _ := lbconfig.LoadClusters(&config, &lg)
	// reflect.DeepEqual(lbclusters, expected) occassionally fails as the array order is not always the same
	// so comparing element par element
	i := 0
	for _, e := range expected {
		for _, c := range lbclusters {
			if c.Cluster_name == e.Cluster_name {
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
}
