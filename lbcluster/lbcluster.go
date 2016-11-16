package lbcluster

import (
	"encoding/json"
	"fmt"
	//"github.com/tiebingzhang/wapsnmp"
	"github.com/k-sone/snmpgo"
	"github.com/miekg/dns"
	"io/ioutil"
	"log/syslog"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const WorstValue int = 99999
const TIMEOUT int = 10
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
	Previous_best_hosts_dns []string
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

func fisher_yates_shuffle(array []string) []string {
	var jval, ival string
	var i, j int32
	for i = int32(len(array) - 1); i > 0; i-- {
		j = rand.Int31n(i + 1)
		if i == j {
			continue
		}
		jval = array[j]
		ival = array[i]
		array[j] = ival
		array[i] = jval
	}
	return array
}

type MetricPolicyApplier interface {
	Apply_metric_minino()
	Apply_metric_minimum()
	Apply_metric_cmsweb()
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (self *LBCluster) Apply_metric_minino() {
	self.write_to_log("Got metric minino = " + self.Parameters.Metric)
	pl := make(PairList, len(self.Host_metric_table))
	i := 0
	for k, v := range self.Host_metric_table {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(pl)
	fmt.Println(pl)
	var sorted_host_list []string
	var useful_host_list []string
	for _, v := range pl {
		if (v.Value > 0) && (v.Value < WorstValue) {
			useful_host_list = append(useful_host_list, v.Key)
		}
		sorted_host_list = append(sorted_host_list, v.Key)
	}
	fmt.Println(useful_host_list)
	useful_hosts := len(useful_host_list)
	list_length := len(pl)
	max := self.Parameters.Best_hosts
	if max == -1 {
		max = list_length
	}
	if max > list_length {
		self.write_to_log(fmt.Sprintf("WARNING: impossible to return %v hosts from the list of %v hosts (%v). Check the configuration of cluster %v. Returning %v hosts.", max, list_length, sorted_host_list, self.Cluster_name, list_length))
		max = list_length
	}
	if list_length == 0 {
		self.write_to_log(fmt.Sprintf("ERROR: cluster %v has no hosts defined ! Check the configuration.", self.Cluster_name))
		self.Current_best_hosts = []string{"unknown"}
	} else if useful_hosts == 0 {
		if self.Parameters.Metric == "minimum" {
			self.write_to_log(fmt.Sprintf("WARNING: no usable hosts found for cluster %v ! Returning random %v hosts.", self.Cluster_name, max))
			sorted_host_list = fisher_yates_shuffle(sorted_host_list)
			self.Current_best_hosts = sorted_host_list[:max]
		} else if (self.Parameters.Metric == "minino") || (self.Parameters.Metric == "cmsweb") {
			self.write_to_log(fmt.Sprintf("WARNING: no usable hosts found for cluster %v ! Returning no hosts.", self.Cluster_name))
			self.Current_best_hosts = useful_host_list
		}
	} else {
		if useful_hosts < max {
			self.write_to_log(fmt.Sprintf("WARNING: only %v useable hosts found in cluster %v", useful_hosts, self.Cluster_name))
			max = useful_hosts
		}
		self.Current_best_hosts = useful_host_list[:max]
		fmt.Println(self.Current_best_hosts)
	}
	return
}

func (self *LBCluster) Apply_metric_minimum() {
	self.write_to_log("Got metric minimum = " + self.Parameters.Metric)
	self.Apply_metric_minino()
	return
}

func (self *LBCluster) Apply_metric_cmsweb() {
	self.write_to_log("Got metric cmsweb = " + self.Parameters.Metric)
	self.Apply_metric_minino()
	return
}

func (self *LBCluster) Externally_visible() bool {
	return self.Parameters.External
}

func (self *LBCluster) Time_to_refresh() bool {
	// self.write_to_log(fmt.Sprintf("Time_of_last_evaluation = %v now = %v Time_of_last_evaluation + polling_int = %v result = %v Cluster_name = %v\n", self.Time_of_last_evaluation, time.Now(), self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval)*time.Second), self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval)*time.Second).After(time.Now()), self.Cluster_name))
	return self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval) * time.Second).Before(time.Now())
}

func (self *LBCluster) write_to_log(msg string) error {
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

func (self *LBCluster) Find_best_hosts() {
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
	//self.Apply_metric_minino()
	reflect.ValueOf(a).MethodByName(methodName).Call([]reflect.Value{})
	self.Time_of_last_evaluation = time.Now()
	none := ""
	if len(self.Current_best_hosts) == 0 {
		none = "NONE"
	}
	self.write_to_log(fmt.Sprintf("best hosts for %v are: %v%v\n", self.Cluster_name, strings.Join(self.Current_best_hosts, " "), none))
}

func (self *LBCluster) evaluate_hosts() {
	result := make(chan RetSnmp, 200)
	for h := range self.Host_metric_table {
		currenthost := h
		self.write_to_log("contacting cluster: " + self.Cluster_name + " node: " + currenthost)
		go self.snmp_req(currenthost, result)
	}
	for range self.Host_metric_table {
		//time.Sleep(1 * time.Millisecond)
		select {
		case metrichostlog := <-result:
			self.Host_metric_table[metrichostlog.Host] = metrichostlog.Metric
			self.write_to_log(metrichostlog.Log)
		}
	}
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

func (self *LBCluster) snmp_req(host string, result chan<- RetSnmp) {
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
	transport := self.transportToUse(host)
	//wapsnmp.DoGetTestV3(host, OID, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "NOPRIV", self.Loadbalancing_password)
	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:       snmpgo.V3,
		Network:       transport,
		Address:       host + ":161",
		Retries:       1,
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
	// could not reproduce the problem on 16 Nov 2016
	// problem does not happen with retries: 0 in SNMPArguments
	for i := 0; i <= 1; i++ {
		if err = snmp.Open(); err != nil {
			// Failed to open connection
			if _, ok := err.(*snmpgo.MessageError); ok {
				snmp.Close()
				logmessage = logmessage + " - " + fmt.Sprintf("open error: %v retrying: %v", err, i)
				continue
			} else {
				logmessage = logmessage + fmt.Sprintf("snmp open %v failed with %v", transport, err)
				logmessage = fmt.Sprintf("contacted  cluster: %v node: %v - %v - %v - setting reply %v", self.Cluster_name, host, transport, logmessage, metric)
				result <- RetSnmp{metric, host, logmessage}
				return
			}
		}

		pdu, err := snmp.GetRequest(oids)
		if err != nil {
			if _, ok := err.(*snmpgo.MessageError); ok {
				snmp.Close()
				logmessage = logmessage + fmt.Sprintf("get error: %v retrying: %v", err, i)
				continue
			} else {
				logmessage = logmessage + fmt.Sprintf("snmp get failed with %v", err)
				logmessage = fmt.Sprintf("contacted  cluster: %v node: %v - %v - %v - setting reply %v", self.Cluster_name, host, transport, logmessage, metric)
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

	if logmessage == "" {
		logmessage = fmt.Sprintf("contacted  cluster: %v node: %v transport: %v - reply was %v", self.Cluster_name, host, transport, metric)
	} else {
		logmessage = logmessage + " - " + fmt.Sprintf("contacted  cluster: %v node: %v transport: %v - reply was %v", self.Cluster_name, host, transport, metric)
	}
	result <- RetSnmp{metric, host, logmessage}
	return
}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func (self *LBCluster) transportToUse(hostname string) string {
	// udp (IPv4) transport
	result := "udp"
	ips, err := net.LookupIP(hostname)
	if err != nil {
		self.write_to_log(fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", hostname, err))
		return result
	}
	for _, ip := range ips {
		// If there is an IPv6 address use udp6 transport
		if ip.To4() == nil {
			result = "udp6"
			break
		}
	}
	return result
}

func (self *LBCluster) Update_dns(keyName, tsigKey, dnsManager string) error {
	pbhDns := strings.Join(self.Previous_best_hosts_dns, " ")
	cbh := strings.Join(self.Current_best_hosts, " ")
	if pbhDns == cbh {
		self.write_to_log(fmt.Sprintf("DNS not update for %v with  keyName %v cbh == pbhDns == %v", self.Cluster_name, keyName, cbh))
		return nil
	}
	cluster_name := self.Cluster_name
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	//best_hosts_len := len(self.Current_best_hosts)
	m := new(dns.Msg)
	m.SetUpdate(cluster_name + ".")
	m.Id = 1234
	rr_removeA, _ := dns.NewRR(cluster_name + ". 60 IN A 127.0.0.1")
	rr_removeAAAA, _ := dns.NewRR(cluster_name + ". 60 IN AAAA ::1")
	m.RemoveRRset([]dns.RR{rr_removeA})
	m.RemoveRRset([]dns.RR{rr_removeAAAA})

	for _, hostname := range self.Current_best_hosts {
		ips, err := net.LookupIP(hostname)
		if err != nil {
			self.write_to_log(fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", hostname, err))
			continue
		}
		for _, ip := range ips {
			var rr_insert dns.RR
			if ip.To4() != nil {
				rr_insert, _ = dns.NewRR(cluster_name + ". 60 IN A " + ip.String())
			} else if ip.To16() != nil {
				rr_insert, _ = dns.NewRR(cluster_name + ". 60 IN AAAA " + ip.String())
			}
			m.Insert([]dns.RR{rr_insert})
		}
	}
	c := new(dns.Client)
	m.SetTsig(keyName, dns.HmacMD5, 300, time.Now().Unix())
	c.TsigSecret = map[string]string{keyName: tsigKey}
	_, _, err := c.Exchange(m, dnsManager+":53")
	if err != nil {
		self.write_to_log(fmt.Sprintf("DNS update for %v failed with (%v)", cluster_name, err))
	}
	self.write_to_log(fmt.Sprintf("DNS update for %v with  keyName %v", cluster_name, keyName))
	return err
}

func (self *LBCluster) Get_state_dns(dnsManager string) error {
	cluster_name := self.Cluster_name
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	m := new(dns.Msg)
	m.SetQuestion(cluster_name+".", dns.TypeANY)
	in, err := dns.Exchange(m, dnsManager+":53")
	if err != nil {
		self.write_to_log(fmt.Sprintf("%v", err))
		return err
	}

	var ips []net.IP
	for _, a := range in.Answer {
		if t, ok := a.(*dns.A); ok {
			fmt.Println(t)
			fmt.Println(t.A)
			ips = append(ips, t.A)
		}
		if t, ok := a.(*dns.AAAA); ok {
			fmt.Println(t)
			fmt.Println(t.AAAA)
			ips = append(ips, t.AAAA)
		}
	}
	//addrs, err := net.LookupHost(cluster_name)
	//ips, err := net.LookupIP(cluster_name)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(ips)
	var name string
	var host_list []string
	for _, ip := range ips {
		names, err := net.LookupAddr(ip.String())
		if err != nil {
			self.write_to_log(fmt.Sprintf("%v", err))
			return err
		}
		if len(names) > 0 {
			if len(names) == 1 {
				name = strings.TrimRight(names[0], ".")
			} else {
				name, err = net.LookupCNAME(names[0])
				if err != nil {
					self.write_to_log(fmt.Sprintf("%v", err))
					return err
				}
				name = strings.TrimRight(name, ".")
			}
			host_list = append(host_list, name)
		}
	}
	RemoveDuplicates(&host_list)
	self.Previous_best_hosts_dns = host_list
	prevBesthostsDns := make([]string, len(self.Previous_best_hosts_dns))
	prevBesthosts := make([]string, len(self.Previous_best_hosts))
	copy(prevBesthostsDns, self.Previous_best_hosts_dns)
	copy(prevBesthosts, self.Previous_best_hosts)
	sort.Strings(prevBesthostsDns)
	sort.Strings(prevBesthosts)
	pbhDns := strings.Join(prevBesthostsDns, " ")
	pbh := strings.Join(prevBesthosts, " ")
	if pbh != "unknown" {
		if pbh != pbhDns {
			self.write_to_log("WARNING: Prev DNS state " + pbhDns + " - Prev local state  " + pbh + " differ")
		}
	}
	return err
}
