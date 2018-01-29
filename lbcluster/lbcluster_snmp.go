package lbcluster

import (
	"encoding/json"
	"fmt"
	"github.com/k-sone/snmpgo"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
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
					logmessage = logmessage + fmt.Sprintf("node: %s - %s - setting reply %v", host, str, metric)
					result <- RetSnmp{metric, host, logmessage}
					return
				}
			} else {
				logmessage = logmessage + fmt.Sprintf("dat[\"appstate\"] not a string for node %s", host)
			}

		}
		metric = WorstValue
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
		self.Slog.Debug(fmt.Sprintf("%v", err))
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
				logmessage = fmt.Sprintf("contacted node: %v - %v - %v - setting reply %v", host, transport, logmessage, metric)
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
				logmessage = fmt.Sprintf("contacted node: %v - %v - %v - setting reply %v", host, transport, logmessage, metric)
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
		logmessage = fmt.Sprintf("contacted node: %v transport: %v - reply was %v", host, transport, metric)
	} else {
		logmessage = logmessage + " - " + fmt.Sprintf("contacted node: %v transport: %v - reply was %v", host, transport, metric)
	}
	result <- RetSnmp{metric, host, logmessage}
	return
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
