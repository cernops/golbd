package lbhost

import (
	//	"encoding/json"
	"fmt"
	//"io/ioutil"
	"github.com/reguero/go-snmplib"
	//"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	//	"net/http"

	//	"sort"
	//	"strings"
	"time"
)

const TIMEOUT int = 10
const OID string = ".1.3.6.1.4.1.96.255.1"

type LBHost struct {
	Cluster_name           string
	Host_name              string
	Host_response_int      int
	Host_response_string   string
	Host_response_error    string
	Loadbalancing_username string
	Loadbalancing_password string
	LogFile                string
	logMu                  sync.Mutex
}

func (self *LBHost) Snmp_req() {

	self.Host_response_int = -100
	/*
		if self.Parameters.Metric == "cmsweb" {
			if message := self.checkRogerState(host); message != "" {
				result <- RetSnmp{-99, host, message}
				return
			}
			metric = WorstValue
		} */
	transport := self.transportToUse()

	snmp, err := snmplib.NewSNMPv3(self.Host_name, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "NOPRIV", self.Loadbalancing_password,
		time.Duration(TIMEOUT)*time.Second, 2)
	if err != nil {
		// Failed to create snmpgo.SNMP object
		self.Host_response_error = fmt.Sprint("contacted node: %v error creating the snmp object: %v", self.Host_name, err)
		return
	}
	defer snmp.Close()
	err = snmp.Discover()

	if err != nil {
		self.Host_response_error = fmt.Sprintf("contacted node: %v error in the snmp discovery: %v", self.Host_name, err)
		return
	}

	oid, err := snmplib.ParseOid(OID)

	if err != nil {
		// Failed to parse Oids
		self.Host_response_error = fmt.Sprintf("contacted node: %v Error parsing the OID %v", self.Host_name, err)
		return
	}
	pdu, err := snmp.GetV3(oid)

	if err != nil {
		self.Host_response_error = fmt.Sprintf("contacted node: %v The getv3 gave the following error: %v ", self.Host_name, err)
		return
	}

	self.Write_to_log("INFO", fmt.Sprintf("contacted node: %v transport: %v - reply was %v", self.Host_name, transport, pdu))

	//var pduInteger int
	switch t := pdu.(type) {
	case int:
		self.Host_response_int = pdu.(int)
	case string:
		self.Host_response_string = pdu.(string)
	default:
		self.Host_response_error = fmt.Sprintf("The node returned an unexpected type %s in %v", t, pdu)
		return
	}
}

func (self *LBHost) transportToUse() string {
	// udp (IPv4) transport
	result := "udp"
	ips, err := net.LookupIP(self.Host_name)
	if err != nil {
		self.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", self.Host_name, err))
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

func (self *LBHost) Write_to_log(level string, msg string) error {
	var err error
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	timestamp := time.Now().Format(time.Stamp)
	msg = fmt.Sprintf("%s lbd[%d]: %s: host: %s %s", timestamp, os.Getpid(), level, self.Host_name, msg)

	self.logMu.Lock()
	defer self.logMu.Unlock()

	f, err := os.OpenFile(self.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, msg)

	return err
}

func (self *LBHost) Get_load_for_alias(cluster_name string) int {
	pduInteger := self.Host_response_int

	re := regexp.MustCompile(cluster_name + "=([0-9]+)")
	submatch := re.FindStringSubmatch(self.Host_response_string)

	if submatch != nil {
		pduInteger, _ = strconv.Atoi(submatch[1])
	}

	return pduInteger

}
