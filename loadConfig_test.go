package main

import (
	"gitlab.cern.ch/lb-experts/golbd/lbcluster"
	//"sync"
	"io/ioutil"
	"reflect"
	"syscall"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "testloadconfig")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(f.Name())
	ioutil.WriteFile(f.Name(), []byte(`#
# Who is the primary master to upload the data ?
#  - fully qualified DNS name
#
master = lbdxyz.cern.ch

#
# Heartbeat details
#
heartbeat_path = /work/go/src/github.com/cernops/golbd
heartbeat_file = heartbeat

#
# TSIG HMAC-MD5 algorithm keys for DNS access
#
tsig_key_prefix = abcd-
tsig_internal_key = xxx123==
tsig_external_key = yyy123==

#
# SNMPv3 password for 'loadbalancing' user
#
snmpd_password = zzz123

#
# Which node manages information in DNS servers ?
#
dns_manager = 111.111.0.111

#
# Clusters definitions
#
parameters test01.cern.ch = behaviour#mindless best_hosts#2 external#yes metric#cmsfrontier polling_interval#6 statistics#long
parameters test02.cern.ch = behaviour#mindless best_hosts#10 external#no metric#cmsfrontier polling_interval#6 statistics#long

clusters test01.cern.ch = lxplus142.cern.ch lxplus177.cern.ch
clusters test02.cern.ch = lxplus013.cern.ch lxplus038.cern.ch lxplus025.cern.ch
`), 0644)
	lg := lbcluster.Log{Syslog: false, Stdout: true, Debugflag: false}
	expected := Config{Master: "lbdxyz.cern.ch",
		HeartbeatFile: "heartbeat",
		HeartbeatPath: "/work/go/src/github.com/cernops/golbd",
		//HeartbeatMu:     sync.Mutex{0, 0},
		TsigKeyPrefix:   "abcd-",
		TsigInternalKey: "xxx123==",
		TsigExternalKey: "yyy123==",
		SnmpPassword:    "zzz123",
		DnsManager:      "111.111.0.111",
		Clusters: map[string][]string{"test01.cern.ch": {"lxplus142.cern.ch", "lxplus177.cern.ch"},
			"test02.cern.ch": {"lxplus013.cern.ch", "lxplus038.cern.ch", "lxplus025.cern.ch"}},
		Parameters: map[string]lbcluster.Params{"test01.cern.ch": {Behaviour: "mindless", Best_hosts: 2, External: true, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"},
			"test02.cern.ch": {Behaviour: "mindless", Best_hosts: 10, External: false, Metric: "cmsfrontier", Polling_interval: 6, Statistics: "long"}}}

	config, e := loadConfig(f.Name(), &lg)
	if e != nil {
		t.Errorf("loadConfig Error: %v", e.Error())
	} else {
		if !reflect.DeepEqual(config, &expected) {
			t.Errorf("loadConfig: got %v expected %v", config, &expected)
		}
	}

}
