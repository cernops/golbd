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
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}
	return lbcluster.LBCluster{Cluster_name: name,
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
		Host_metric_table: map[string]lbcluster.Node{
			"lxplus132.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus041.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus130.cern.ch":               lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"monit-kafkax-17be060b0d.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}}},
		Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_ips:      []net.IP{},
		Previous_best_ips_dns: []net.IP{},
		Slog:                  &lg,
		Current_index:         0}
}

func getSecondTestCluster() lbcluster.LBCluster {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}
	return lbcluster.LBCluster{Cluster_name: "test02.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "zzz123",
		Host_metric_table: map[string]lbcluster.Node{
			"lxplus013.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus038.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}},
			"lxplus025.cern.ch": lbcluster.Node{Load: 100000, IPs: []net.IP{}}},
		Parameters: lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_ips:      []net.IP{},
		Previous_best_ips_dns: []net.IP{},
		Slog:                  &lg,
		Current_index:         0}
}
func getHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.LBHost {
	hostsToCheck := map[string]lbhost.LBHost{
		"lxplus132.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name: "lxplus132.cern.ch",
			Host_transports: []lbhost.LBHostTransportResult{
				lbhost.LBHostTransportResult{Transport: "udp6", Response_int: 2, Response_string: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), Response_error: ""},
				lbhost.LBHostTransportResult{Transport: "udp", Response_int: 2, Response_string: "", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
			},
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"lxplus041.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name: "lxplus041.cern.ch",
			Host_transports: []lbhost.LBHostTransportResult{
				lbhost.LBHostTransportResult{Transport: "udp6", Response_int: 3, Response_string: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), Response_error: ""},
				lbhost.LBHostTransportResult{Transport: "udp", Response_int: 3, Response_string: "", IP: net.ParseIP("188.184.116.81"), Response_error: ""},
			},
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"lxplus130.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "lxplus130.cern.ch",
			Host_transports:        []lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{Transport: "udp", Response_int: 27, Response_string: "", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "monit-kafkax-17be060b0d.cern.ch",
			Host_transports:        []lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{Transport: "udp", Response_int: 100000, Response_string: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
	}

	return hostsToCheck
}
func getBadHostsToCheck(c lbcluster.LBCluster) map[string]lbhost.LBHost {
        badHostsToCheck := map[string]lbhost.LBHost{
                "lxplus132.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
                        Host_name: "lxplus132.cern.ch",
                        Host_transports: []lbhost.LBHostTransportResult{
                                lbhost.LBHostTransportResult{Transport: "udp6", Response_int: -2, Response_string: "", IP: net.ParseIP("2001:1458:d00:2c::100:a6"), Response_error: ""},
                                lbhost.LBHostTransportResult{Transport: "udp", Response_int: -2, Response_string: "", IP: net.ParseIP("188.184.108.98"), Response_error: ""},
                        },
                        Loadbalancing_username: c.Loadbalancing_username,
                        Loadbalancing_password: c.Loadbalancing_password,
                        LogFile:                c.Slog.TofilePath,
                        Debugflag:              c.Slog.Debugflag,
                },
                "lxplus041.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
                        Host_name: "lxplus041.cern.ch",
                        Host_transports: []lbhost.LBHostTransportResult{
                                lbhost.LBHostTransportResult{Transport: "udp6", Response_int: -3, Response_string: "", IP: net.ParseIP("2001:1458:d00:32::100:51"), Response_error: ""},
                                lbhost.LBHostTransportResult{Transport: "udp", Response_int: -3, Response_string: "", IP: net.ParseIP("188.184.116.81"), Response_error: ""},
                        },
                        Loadbalancing_username: c.Loadbalancing_username,
                        Loadbalancing_password: c.Loadbalancing_password,
                        LogFile:                c.Slog.TofilePath,
                        Debugflag:              c.Slog.Debugflag,
                },
                "lxplus130.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
                        Host_name:              "lxplus130.cern.ch",
                        Host_transports:        []lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{Transport: "udp", Response_int: -27, Response_string: "", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
                        Loadbalancing_username: c.Loadbalancing_username,
                        Loadbalancing_password: c.Loadbalancing_password,
                        LogFile:                c.Slog.TofilePath,
                        Debugflag:              c.Slog.Debugflag,
                },
                "monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
                        Host_name:              "monit-kafkax-17be060b0d.cern.ch",
                        Host_transports:        []lbhost.LBHostTransportResult{lbhost.LBHostTransportResult{Transport: "udp", Response_int: 100000, Response_string: "monit-kafkax.cern.ch=816,monit-kafka.cern.ch=816,test01.cern.ch=816", IP: net.ParseIP("188.184.108.100"), Response_error: ""}},
                        Loadbalancing_username: c.Loadbalancing_username,
                        Loadbalancing_password: c.Loadbalancing_password,
                        LogFile:                c.Slog.TofilePath,
                        Debugflag:              c.Slog.Debugflag,
                },
        }

        return badHostsToCheck
}
func getHost(hostname string, responseInt int, responseString string) lbhost.LBHost {

	return lbhost.LBHost{Cluster_name: "test01.cern.ch",
		Host_name: hostname,
		Host_transports: []lbhost.LBHostTransportResult{
			lbhost.LBHostTransportResult{Transport: "udp", Response_int: responseInt, Response_string: responseString, IP: net.ParseIP("188.184.108.98"), Response_error: ""}},
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "XXXX",
		LogFile:                "",
		Debugflag:              false,
	}

}
func TestLoadClusters(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

	config := lbconfig.Config{Master: "lbdxyz.cern.ch",
		HeartbeatFile: "heartbeat",
		HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
		//HeartbeatMu:     sync.Mutex{0, 0},
		TsigKeyPrefix:   "abcd-",
		TsigInternalKey: "xxx123==",
		TsigExternalKey: "yyy123==",
		SnmpPassword:    "zzz123",
		DNSManager:      "111.111.0.111",
		Clusters:        map[string][]string{"test01.cern.ch": {"lxplus132.cern.ch", "lxplus041.cern.ch", "lxplus130.cern.ch", "monit-kafkax-17be060b0d.cern.ch"}, "test02.cern.ch": {"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus025.cern.ch"}},
		Parameters: map[string]lbcluster.Params{"test01.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 2,
			External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			"test02.cern.ch": lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"}}}
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
