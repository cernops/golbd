package lbcluster

import (
	"encoding/json"
	"fmt"
	//"github.com/tiebingzhang/wapsnmp"
	"github.com/k-sone/snmpgo"
	"io/ioutil"
	"log/syslog"
	//"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const WorstValue int = 99999
const TIMEOUT int = 5
const OID string = ".1.3.6.1.4.1.96.255.1"

type LBCluster struct {
	Cluster_name            string
	Loadbalancing_username  string
	Loadbalancing_password  string
	Host_metric_table       map[string]int
	Parameters              Params
	Time_of_last_evaluation time.Time
	Current_best_hosts      []string
	Previous_best_hosts     []string
	Statistics_filename     string
	Per_cluster_filename    string
	Current_index           int
	Slog                    Log
}

type Params struct {
	Behaviour        string
	Best_hosts       int
	External         bool
	Metric           string
	Polling_interval int
	Statistics       string
}

type RetSnmp struct {
	Metric int
	Host   string
	Log    string
}

type Log struct {
	Writer syslog.Writer
	Debug  bool
}

type Logger interface {
	Info(s string) error
	Warning(s string) error
}

func (l Log) Info(s string) error {
	err := l.Writer.Info(s)
	if l.Debug {
		fmt.Println(s)
	}
	return err

}

func (l Log) Warning(s string) error {
	err := l.Writer.Warning(s)
	if l.Debug {
		fmt.Println(s)
	}
	return err

}

type MetricPolicyApplier interface {
	Apply_metric_minino()
	Apply_metric_minimum()
	Apply_metric_cmsweb()
}

func (self LBCluster) Apply_metric_minino() {
	self.write_to_log("Got metric minino = " + self.Parameters.Metric)
	return
}

func (self LBCluster) Apply_metric_minimum() {
	self.write_to_log("Got metric minimum = " + self.Parameters.Metric)
	return
}

func (self LBCluster) Apply_metric_cmsweb() {
	self.write_to_log("Got metric cmsweb = " + self.Parameters.Metric)
	return
}

func (self LBCluster) Time_to_refresh() bool {
	if self.Time_of_last_evaluation.IsZero() {
		return true
	} else {
		return self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval) * time.Second).After(time.Now())
	}
}

func (self LBCluster) write_to_log(msg string) error {
	self.Slog.Info(msg)
	f, err := os.OpenFile(self.Per_cluster_filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	tag := "lbd"
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}
	timestamp := time.Now().Format(time.Stamp)
	_, err = fmt.Fprintf(f, "%s %s[%d]: %s%s",
		timestamp,
		tag, os.Getpid(), msg, nl)
	return err
}

func (self LBCluster) Find_best_hosts() {
	self.Previous_best_hosts = self.Current_best_hosts
	self.evaluate_hosts()
	methodName := "Apply_metric_" + self.Parameters.Metric
	var a MetricPolicyApplier
	a = self
	_, ok := reflect.TypeOf(a).MethodByName(methodName)
	if !ok {
		self.write_to_log("ERROR: wrong parameter(metric) in definition of cluster " + self.Parameters.Metric)
	}
	// invoke m
	self.write_to_log(self.Cluster_name + " invoking " + self.Parameters.Metric)
	reflect.ValueOf(a).MethodByName(methodName).Call([]reflect.Value{})
}

func (self LBCluster) evaluate_hosts() {
	var wg sync.WaitGroup
	result := make(chan RetSnmp, 200)
	for h := range self.Host_metric_table {
		self.write_to_log("contacting cluster: " + self.Cluster_name + " node: " + h)
		wg.Add(1)
		go self.snmp_req(h, &wg, result)
	}
	for range self.Host_metric_table {
		time.Sleep(1 * time.Millisecond)
		select {
		case metrichostlog := <-result:
			self.Host_metric_table[metrichostlog.Host] = metrichostlog.Metric
			self.write_to_log(metrichostlog.Log)
		}
	}
	wg.Wait()
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

func (self LBCluster) snmp_req(host string, wg *sync.WaitGroup, result chan<- RetSnmp) {
	defer wg.Done()
	//time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
	metric := -100
	logmessage := ""
	if self.Parameters.Metric == "cmsweb" {
		connectTimeout := (10 * time.Second)
		readWriteTimeout := (20 * time.Second)
		httpClient := NewTimeoutClient(connectTimeout, readWriteTimeout)
		response, err := httpClient.Get("http://woger-direct.cern.ch:9098/roger/v1/state/" + host)
		if err != nil {
			logmessage = logmessage + fmt.Sprintf("%s", err)
		} else {
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				logmessage = logmessage + fmt.Sprintf("%s", err)
			}
			var dat map[string]interface{}
			if err := json.Unmarshal([]byte(contents), &dat); err != nil {
				logmessage = logmessage + " - " + fmt.Sprintf("%s", host)
				logmessage = logmessage + " - " + fmt.Sprintf("%v", response.Body)
				logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
			}
			if str, ok := dat["appstate"].(string); ok {
				if str != "production" {
					metric = -99
					logmessage = logmessage + fmt.Sprintf("cluster: %s node: %s - %s - setting reply %v", self.Cluster_name, host, str, metric)
					result <- RetSnmp{metric, host, logmessage}
					return
				}
			} else {
				logmessage = logmessage + fmt.Sprintf("dat[\"appstate\"] not a string for node %s", host)
			}

		}
		metric = WorstValue - 1
	}
	//wapsnmp.DoGetTestV3(host, OID, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "NOPRIV", self.Loadbalancing_password)
	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:       snmpgo.V3,
		Address:       host + ":161",
		Retries:       0,
		UserName:      self.Loadbalancing_username,
		SecurityLevel: snmpgo.AuthNoPriv,
		AuthProtocol:  snmpgo.Md5,
		AuthPassword:  self.Loadbalancing_password,
		Timeout:       time.Duration(TIMEOUT) * time.Second,
	})
	if err != nil {
		// Failed to create snmpgo.SNMP object
		fmt.Println(err)
		logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
		result <- RetSnmp{metric, host, logmessage}
		return
	}

	oids, err := snmpgo.NewOids([]string{
		OID,
	})
	if err != nil {
		// Failed to parse Oids
		logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
		result <- RetSnmp{metric, host, logmessage}
		return
	}
	// retry MessageId mismatch
	// although problem should not happen wiht retries: 0 in SNMPArguments
	for i := 0; i <= 1; i++ {
		if err = snmp.Open(); err != nil {
			// Failed to open connection
			if _, ok := err.(*snmpgo.MessageError); ok {
				snmp.Close()
				logmessage = logmessage + " - " + fmt.Sprintf("retrying: %v", i)
				continue
			} else {
				logmessage = logmessage + fmt.Sprintf("snmp open failed with %v", err)
				logmessage = fmt.Sprintf("contacted  cluster: %v node: %v - %v - setting reply %v", self.Cluster_name, host, logmessage, metric)
				result <- RetSnmp{metric, host, logmessage}
				return
			}
		}

		pdu, err := snmp.GetRequest(oids)
		if err != nil {
			if _, ok := err.(*snmpgo.MessageError); ok {
				snmp.Close()
				logmessage = logmessage + fmt.Sprintf("retrying: %v", i)
				continue
			} else {
				logmessage = logmessage + fmt.Sprintf("snmp get failed with %v", err)
				logmessage = fmt.Sprintf("contacted  cluster: %v node: %v - %v - setting reply %v", self.Cluster_name, host, logmessage, metric)
				result <- RetSnmp{metric, host, logmessage}
				return
			}
		}
		if pdu.ErrorStatus() != snmpgo.NoError {
			// Received an error from the agent
			logmessage = logmessage + " - " + fmt.Sprintf("%v %v", pdu.ErrorStatus(), pdu.ErrorIndex())
		}

		// select a VarBind
		Varbind := pdu.VarBinds().MatchOid(oids[0])
		if Varbind.Variable.Type() == "Integer" {
			metricstr := Varbind.Variable.String()
			if metric, err = strconv.Atoi(metricstr); err != nil {
				logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
				result <- RetSnmp{metric, host, logmessage}
				return
			}
		} else if Varbind.Variable.Type() == "OctetString" {
			cskvpair := Varbind.Variable.String()
			kvpair := strings.Split(cskvpair, ",")
			for _, kv := range kvpair {
				cm := strings.Split(kv, "=")
				if cm[0] == self.Cluster_name {
					if metric, err = strconv.Atoi(cm[1]); err != nil {
						logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
						result <- RetSnmp{metric, host, logmessage}
						return
					}
				}
			}
		}
		break
	}
	defer snmp.Close()

	logmessage = logmessage + "\n" + fmt.Sprintf("contacted  cluster: %v node: %v - reply was %v", self.Cluster_name, host, metric)
	result <- RetSnmp{metric, host, logmessage}
	return
}
