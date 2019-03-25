package main

import (
	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestFindBestHosts(t *testing.T) {
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}

	c := lbcluster.LBCluster{Cluster_name: "test01.cern.ch",
		Loadbalancing_username: "loadbalancing",
		Loadbalancing_password: "OidLvSu8amNMRz258Udy74tO60p47n0RA4RzaT3j2hhnJkEQg9",
		Host_metric_table:      map[string]int{"lxplus132.cern.ch": 100000, "lxplus041.cern.ch": 100000, "lxplus130.cern.ch": 100000, "monit-kafkax-17be060b0d.cern.ch": 100000},
		Parameters:             lbcluster.Params{Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 300, Statistics: "long"},
		//Time_of_last_evaluation time.Time
		Current_best_hosts:      []string{"unknown"},
		Previous_best_hosts:     []string{"unknown"},
		Previous_best_hosts_dns: []string{"unknown"},
		Slog:          &lg,
		Current_index: 0}
	hosts_to_check := map[string]lbhost.LBHost{
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
	expected_host_metric_table := map[string]int{"monit-kafkax-17be060b0d.cern.ch": 816, "lxplus132.cern.ch": 2, "lxplus041.cern.ch": 3, "lxplus130.cern.ch": 27}
	expected_previous_best_hosts := c.Current_best_hosts
	expected_current_best_hosts := []string{"lxplus041.cern.ch", "lxplus132.cern.ch"}

	c.Find_best_hosts(hosts_to_check)
	if !reflect.DeepEqual(c.Host_metric_table, expected_host_metric_table) {
		t.Errorf("e.Find_best_hosts: c.Host_metric_table: got\n%v\nexpected\n%v", c.Host_metric_table, expected_host_metric_table)
	}
	if !reflect.DeepEqual(c.Previous_best_hosts, expected_previous_best_hosts) {
		t.Errorf("e.Find_best_hosts: c.Previous_best_hosts: got\n%v\nexpected\n%v", c.Previous_best_hosts, expected_previous_best_hosts)
	}
	if !reflect.DeepEqual(c.Current_best_hosts, expected_current_best_hosts) {
		t.Errorf("e.Find_best_hosts: c.Current_best_hosts: got\n%v\nexpected\n%v", c.Current_best_hosts, expected_current_best_hosts)
	}
	if c.Time_of_last_evaluation.Add(time.Duration(2) * time.Second).Before(time.Now()) {
		t.Errorf("e.Find_best_hosts: c.Time_of_last_evaluation: got\n%v\ncurrent time\n%v", c.Time_of_last_evaluation, time.Now())
	}
}
