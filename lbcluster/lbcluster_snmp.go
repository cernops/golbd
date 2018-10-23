package lbcluster

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/reguero/go-snmplib"
)

//This one has only internal methods. They should not be called from outside the lbcluster

type RetSnmp struct {
	Metric int
	Host   string
	Log    string
}

func (self *LBCluster) evaluate_hosts() {
	result := make(chan RetSnmp, 200)
	for currenthost := range self.Host_metric_table {
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

	snmp, err := snmplib.NewSNMPv3(host, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "NOPRIV", self.Loadbalancing_password,
		time.Duration(TIMEOUT)*time.Second, 2)
	if err != nil {
		// Failed to create snmpgo.SNMP object
		result <- RetSnmp{metric, host, fmt.Sprintf("contacted node: %v error creating the snmp object: %v", host, err)}
		return
	}
	defer snmp.Close()
	err = snmp.Discover()

	if err != nil {
		result <- RetSnmp{metric, host, fmt.Sprintf("contacted node: %v error in the snmp discovery of ", host)}
		return
	}

	oid, err := snmplib.ParseOid(OID)

	if err != nil {
		// Failed to parse Oids
		result <- RetSnmp{metric, host, fmt.Sprintf("contacted node: %v Error parsing the OID %v", host, err)}
		return
	}
	pdu, err := snmp.GetV3(oid)

	if err != nil {
		result <- RetSnmp{metric, host, fmt.Sprintf("contacted node: %v The getv3 gave the following error: %v ", host, err)}
		return
	}

	logmessage := fmt.Sprintf("contacted node: %v transport: %v - reply was %v", host, transport, pdu)

	var pduInteger int
	switch t := pdu.(type) {
	case int:
		pduInteger = pdu.(int)
	case string:
		re := regexp.MustCompile(self.Cluster_name + "=([0-9]+)")
		submatch := re.FindStringSubmatch(pdu.(string))
		if submatch != nil {
			pduInteger, err = strconv.Atoi(submatch[1])
		}
	default:
		result <- RetSnmp{metric, host, fmt.Sprintf("The node returned an unexpected type %s in %v", t, pdu)}
		return
	}
	result <- RetSnmp{pduInteger, host, logmessage}
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
