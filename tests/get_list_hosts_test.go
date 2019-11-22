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
		"lxplus041.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "lxplus041.cern.ch",
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"monit-kafkax-17be060b0d.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "monit-kafkax-17be060b0d.cern.ch",
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"lxplus132.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "lxplus132.cern.ch",
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
		"lxplus130.cern.ch": lbhost.LBHost{Cluster_name: c.Cluster_name,
			Host_name:              "lxplus130.cern.ch",
			Loadbalancing_username: c.Loadbalancing_username,
			Loadbalancing_password: c.Loadbalancing_password,
			LogFile:                c.Slog.TofilePath,
			Debugflag:              c.Slog.Debugflag,
		},
	}

	hosts_to_check := make(map[string]lbhost.LBHost)
	c.Get_list_hosts(hosts_to_check)
	if !reflect.DeepEqual(hosts_to_check, expected) {
		t.Errorf("e.Get_list_hosts: got\n%v\nexpected\n%v", hosts_to_check, expected)
	}
}

func TestGetListHostsTwo(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

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
		"lxplus142.cern.ch": lbhost.LBHost{Cluster_name: "test01.cern.ch",
			Host_name:              "lxplus142.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			LogFile:                "",
			Debugflag:              false,
		},
		"lxplus177.cern.ch": lbhost.LBHost{Cluster_name: "test01.cern.ch,test02.cern.ch",
			Host_name:              "lxplus177.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			LogFile:                "",
			Debugflag:              false,
		},
		"lxplus013.cern.ch": lbhost.LBHost{Cluster_name: "test02.cern.ch",
			Host_name:              "lxplus013.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			LogFile:                "",
			Debugflag:              false,
		},
		"lxplus025.cern.ch": lbhost.LBHost{Cluster_name: "test02.cern.ch",
			Host_name:              "lxplus025.cern.ch",
			Loadbalancing_username: "loadbalancing",
			Loadbalancing_password: "zzz123",
			LogFile:                "",
			Debugflag:              false,
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
