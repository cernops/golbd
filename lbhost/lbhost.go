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

type LBHostTransportResult struct {
	Transport       string
	IP              net.IP
	Response_int    int
	Response_string string
	Response_error  string
}
type LBHost struct {
	Cluster_name           string
	Host_name              string
	Host_transports        []LBHostTransportResult
	Loadbalancing_username string
	Loadbalancing_password string
	LogFile                string
	logMu                  sync.Mutex
	Debugflag              bool
}

// todo: refractor into smaller functions
func (self *LBHost) Snmp_req() {

	self.find_transports()

	for i, my_transport := range self.Host_transports {
		my_transport.Response_int = 100000
		transport := my_transport.Transport
		node_ip := my_transport.IP.String()
		/* There is no need to put square brackets around the ipv6 addresses*/
		self.Write_to_log("DEBUG", "Checking the host "+node_ip+" with "+transport)
		snmp, err := snmplib.NewSNMPv3(node_ip, self.Loadbalancing_username, "MD5", self.Loadbalancing_password, "NOPRIV", self.Loadbalancing_password,
			time.Duration(TIMEOUT)*time.Second, 2)
		if err != nil {
			// Failed to create snmpgo.SNMP object
			my_transport.Response_error = fmt.Sprintf("contacted node: error creating the snmp object: %v", err)
		} else {
			defer snmp.Close()
			err = snmp.Discover()

			if err != nil {
				my_transport.Response_error = fmt.Sprintf("contacted node: error in the snmp discovery: %v", err)

			} else {

				oid, err := snmplib.ParseOid(OID)

				if err != nil {
					// Failed to parse Oids
					my_transport.Response_error = fmt.Sprintf("contacted node: Error parsing the OID %v", err)

				} else {
					pdu, err := snmp.GetV3(oid)

					if err != nil {
						my_transport.Response_error = fmt.Sprintf("contacted node: The getv3 gave the following error: %v ", err)

					} else {

						self.Write_to_log("INFO", fmt.Sprintf("contacted node: transport: %v ip: %v - reply was %v", transport, node_ip, pdu))

						//var pduInteger int
						switch t := pdu.(type) {
						case int:
							my_transport.Response_int = pdu.(int)
						case string:
							my_transport.Response_string = pdu.(string)
						default:
							my_transport.Response_error = fmt.Sprintf("The node returned an unexpected type %s in %v", t, pdu)
						}
					}
				}
			}
		}
		self.Host_transports[i] = my_transport

	}

	self.Write_to_log("DEBUG", "All the ips have been tested")
	/*for _, my_transport := range self.Host_transports {
		self.Write_to_log("INFO", fmt.Sprintf("%v", my_transport))
	}*/
}

func (self *LBHost) Write_to_log(level string, msg string) error {
	var err error
	if level == "DEBUG" && !self.Debugflag {
		//The debug messages should not appear
		return nil
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	timestamp := time.Now().Format(time.StampMilli)
	msg = fmt.Sprintf("%s lbd[%d]: %s: cluster: %s node: %s %s", timestamp, os.Getpid(), level, self.Cluster_name, self.Host_name, msg)

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
// todo: instead of polling try adhoc webhook updates
func (self *LBHost) Get_load_for_alias(cluster_name string) int {

	my_load := -200
	for _, my_transport := range self.Host_transports {
		pduInteger := my_transport.Response_int

		re := regexp.MustCompile(cluster_name + "=([0-9]+)")
		submatch := re.FindStringSubmatch(my_transport.Response_string)

		if submatch != nil {
			pduInteger, _ = strconv.Atoi(submatch[1])
		}

		if (pduInteger > 0 && pduInteger < my_load) || (my_load < 0) {
			my_load = pduInteger
		}
		self.Write_to_log("DEBUG", fmt.Sprintf("Possible load is %v", pduInteger))

	}
	self.Write_to_log("DEBUG", fmt.Sprintf("THE LOAD IS %v, ", my_load))

	return my_load
}

func (self *LBHost) Get_working_IPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range self.Host_transports {
		if (my_transport.Response_int > 0) && (my_transport.Response_error == "") {
			my_ips = append(my_ips, my_transport.IP)
		}

	}
	self.Write_to_log("INFO", fmt.Sprintf("The ips for this host are %v", my_ips))
	return my_ips, nil
}

func (self *LBHost) Get_all_IPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range self.Host_transports {
		my_ips = append(my_ips, my_transport.IP)
	}
	self.Write_to_log("INFO", fmt.Sprintf("All ips for this host are %v", my_ips))
	return my_ips, nil
}

func (self *LBHost) Get_Ips() ([]net.IP, error) {

	var ips []net.IP

	var err error

	re := regexp.MustCompile(".*no such host")

	net.DefaultResolver.StrictErrors = true

	for i := 0; i < 3; i++ {
		self.Write_to_log("INFO", "Getting the ip addresses")
		ips, err = net.LookupIP(self.Host_name)
		if err == nil {
			return ips, nil
		}
		self.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v) ", self.Host_name, err))
		submatch := re.FindStringSubmatch(err.Error())
		if submatch != nil {
			self.Write_to_log("INFO", "There is no need to retry this error")
			return nil, err
		}
	}

	self.Write_to_log("ERROR", "After several retries, we couldn't get the ips!. Let's try with partial results")
	net.DefaultResolver.StrictErrors = false
	ips, err = net.LookupIP(self.Host_name)
	if err != nil {
		self.Write_to_log("ERROR", fmt.Sprintf("It didn't work :(. This node will be ignored during this evaluation: %v", err))
	}
	return ips, err
}

func (self *LBHost) find_transports() {
	self.Write_to_log("DEBUG", "Let's find the ips behind this host")

	ips, _ := self.Get_Ips()
	for _, ip := range ips {
		transport := "udp"
		// If there is an IPv6 address use udp6 transport
		if ip.To4() == nil {
			transport = "udp6"
		}
		self.Host_transports = append(self.Host_transports, LBHostTransportResult{Transport: transport,
			Response_int: 100000, Response_string: "", IP: ip,
			Response_error: ""})
	}

}
