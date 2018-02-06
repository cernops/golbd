package lbcluster

import (
	"encoding/json"
	"fmt"
	//"github.com/k-sone/snmpgo"
	"github.com/reguero/go-snmplib"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	//"strings"
	"time"
)

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

//This one has only internal methods. They should not be called from outside the lbcluster

type RetSnmp struct {
	Metric int
	Host   string
	Log    string
}

func (self *LBCluster) evaluate_hosts() {
	result := make(chan RetSnmp, 200)
	for h := range self.Host_metric_table {
		currenthost := h
		self.Write_to_log("INFO", "contacting node: "+currenthost)
		go self.snmp_req(currenthost, result)
	}
	for range self.Host_metric_table {
		//time.Sleep(1 * time.Millisecond)
		select {
		case metrichostlog := <-result:
			self.Host_metric_table[metrichostlog.Host] = metrichostlog.Metric
			self.Write_to_log("INFO", metrichostlog.Log)
		}
	}
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func (self *LBCluster) checkRogerState(host string) string {

	logmessage := ""

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
				return fmt.Sprintf("node: %s - %s - setting reply -99", host, str)
			}
		} else {
			logmessage = logmessage + fmt.Sprintf("dat[\"appstate\"] not a string for node %s", host)
		}
	}
	return logmessage

}

func (self *LBCluster) snmp_req(host string, result chan<- RetSnmp) {

	metric := -100

	if self.Parameters.Metric == "cmsweb" {
		if message := self.checkRogerState(host); message != "" {
			result <- RetSnmp{-99, host, message}
			return
		}
		metric = WorstValue
	}
	transport := self.transportToUse(host)

	snmp, err := snmplib.NewSNMPv3(host, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "DES", self.Loadbalancing_password,
		time.Duration(TIMEOUT)*time.Second, 2)
	if err != nil {
		// Failed to create snmpgo.SNMP object
		result <- RetSnmp{metric, host, "Error creating the snmp object" + fmt.Sprintf("%v", err)}
		return
	}
	defer snmp.Close()
	err = snmp.Discover()

	if err != nil {
		result <- RetSnmp{metric, host, "Error in the snmp discovery of " + host}
		return
	}

	oid, err := snmplib.ParseOid(OID)

	if err != nil {
		// Failed to parse Oids
		result <- RetSnmp{metric, host, "Error parsing the OID " + fmt.Sprintf("%v", err)}
		return
	}
	pdu, err := snmp.GetV3(oid)

	if err != nil {
		result <- RetSnmp{metric, host, "The Getv3 failed! " + fmt.Sprintf("get error: %v ", err)}
		return
	}
	// select a VarBind
	pduString := fmt.Sprintf("%v", pdu)

	logmessage := fmt.Sprintf("contacted node: %v transport: %v - reply was %v", host, transport, pdu)

	if pduInteger, err := strconv.Atoi(pduString); err != nil {
		// THIS MIGHT BE A COMMA SEPARATED LIST
		logmessage = logmessage + " - " + fmt.Sprintf("%v", err)
		result <- RetSnmp{pduInteger, host, logmessage}
		return
	} else {
		result <- RetSnmp{pduInteger, host, logmessage}
		return
	}

}
func (self *LBCluster) transportToUse(hostname string) string {
	// udp (IPv4) transport
	result := "udp"
	ips, err := net.LookupIP(hostname)
	if err != nil {
		self.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", hostname, err))
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
