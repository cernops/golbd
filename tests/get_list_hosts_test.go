package main_test

import (
	"net"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

func TestGetListHostsOne(t *testing.T) {
	c := getTestCluster("test01.cern.ch")

	expected := map[string]lbhost.LBHost{
		"lxplus041.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:   "lxplus041.cern.ch",
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:   "monit-kafkax-17be060b0d.cern.ch",
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus132.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:   "lxplus132.cern.ch",
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus133.subdo.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:   "lxplus133.subdo.cern.ch",
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
		"lxplus130.cern.ch": lbhost.LBHost{ClusterName: c.Cluster_name,
			HostName:   "lxplus130.cern.ch",
			LBUsername: c.Loadbalancing_username,
			LBPassword: c.Loadbalancing_password,
			LogFile:    c.Slog.TofilePath,
			DebugFlag:  c.Slog.Debugflag,
		},
	}

	hosts_to_check := make(map[string]lbhost.LBHost)
	c.Get_list_hosts(hosts_to_check)
	if !reflect.DeepEqual(hosts_to_check, expected) {
		t.Errorf("e.Get_list_hosts: got\n%v\nexpected\n%v", hosts_to_check, expected)
	}
}

func TestGetListHostsTwo(t *testing.T) {
	lg := lbcluster.Log{Stdout: true, Debugflag: false}

	clusters := []lbcluster.LBCluster{
		{Cluster_name: "test01.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			Host_metric_table:      map[string]lbcluster.Node{"lxplus142.cern.ch": lbcluster.Node{}, "lxplus177.cern.ch": lbcluster.Node{}},
			Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_ips: []net.IP{},

			Previous_best_ips_dns: []net.IP{},
			Slog:                  &lg,
			Current_index:         0},
		lbcluster.LBCluster{Cluster_name: "test02.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			Host_metric_table:      map[string]lbcluster.Node{"lxplus013.cern.ch": lbcluster.Node{}, "lxplus177.cern.ch": lbcluster.Node{}, "lxplus025.cern.ch": lbcluster.Node{}},
			Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			//Time_of_last_evaluation time.Time
			Current_best_ips:      []net.IP{},
			Previous_best_ips_dns: []net.IP{},
			Slog:                  &lg,
			Current_index:         0}}

	expected := map[string]lbhost.LBHost{
		"lxplus142.cern.ch": lbhost.LBHost{ClusterName: "test01.cern.ch",
			HostName:   "lxplus142.cern.ch",
			LBUsername: "loadbalancing",
			LBPassword: "zzz123",
			LogFile:    "",
			DebugFlag:  false,
		},
		"lxplus177.cern.ch": lbhost.LBHost{ClusterName: "test01.cern.ch,test02.cern.ch",
			HostName:   "lxplus177.cern.ch",
			LBUsername: "loadbalancing",
			LBPassword: "zzz123",
			LogFile:    "",
			DebugFlag:  false,
		},
		"lxplus013.cern.ch": lbhost.LBHost{ClusterName: "test02.cern.ch",
			HostName:   "lxplus013.cern.ch",
			LBUsername: "loadbalancing",
			LBPassword: "zzz123",
			LogFile:    "",
			DebugFlag:  false,
		},
		"lxplus025.cern.ch": lbhost.LBHost{ClusterName: "test02.cern.ch",
			HostName:   "lxplus025.cern.ch",
			LBUsername: "loadbalancing",
			LBPassword: "zzz123",
			LogFile:    "",
			DebugFlag:  false,
		},
	}

	hosts_to_check := make(map[string]lbhost.LBHost)
	for _, c := range clusters {
		c.Get_list_hosts(hosts_to_check)
	}
	if !reflect.DeepEqual(hosts_to_check, expected) {
		t.Errorf("e.Get_list_hosts: got\n%v\nexpected\n%v", hosts_to_check, expected)
	}
}
