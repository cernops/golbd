package lbhost

import (
	//	"encoding/json"
	"fmt"
	"lb-experts/golbd/lbcluster"

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

const (
	TIMEOUT            int    = 10
	OID                string = ".1.3.6.1.4.1.96.255.1"
	DefaultResponseInt        = 100000
)

type LBHostTransportResult struct {
	Transport       string
	IP              net.IP
	Response_int    int
	Response_string string
	Response_error  string
}
type LBHost struct {
	ClusterConfig  lbcluster.Config
	Host_name      string
	HostTransports []LBHostTransportResult
	Logger         lbcluster.Logger
}

type Host interface {
	GetName() string
	SNMPDiscovery()
	GetClusterConfig() *lbcluster.Config
	GetLoadForAlias(clusterName string) int
	GetWorkingIPs() ([]net.IP, error)
	GetAllIPs() ([]net.IP, error)
	GetIps() ([]net.IP, error)
}

func NewLBHost(clusterConfig lbcluster.Config, logger lbcluster.Logger) Host {
	return &LBHost{
		ClusterConfig: clusterConfig,
		Logger:        logger,
	}
}

func (lh *LBHost) GetName() string {
	return lh.Host_name
}
func (lh *LBHost) GetClusterConfig() *lbcluster.Config {
	return &lh.ClusterConfig
}

// todo: refractor into smaller functions
func (lh *LBHost) SNMPDiscovery() {
	lh.find_transports()
	for i, hostTransport := range lh.HostTransports {
		hostTransport.Response_int = DefaultResponseInt
		node_ip := hostTransport.IP.String()
		/* There is no need to put square brackets around the ipv6 addresses*/
		lh.Write_to_log("DEBUG", "Checking the host "+node_ip+" with "+hostTransport.Transport)
		snmp, err := snmplib.NewSNMPv3(node_ip, lh.ClusterConfig.Loadbalancing_username, "MD5", lh.ClusterConfig.Loadbalancing_password, "NOPRIV", lh.ClusterConfig.Loadbalancing_password,
			time.Duration(TIMEOUT)*time.Second, 2)
		if err != nil {
			hostTransport.Response_error = fmt.Sprintf("contacted node: error creating the snmp object: %v", err)
		} else {
			defer snmp.Close()
			err = snmp.Discover()
			if err != nil {
				hostTransport.Response_error = fmt.Sprintf("contacted node: error in the snmp discovery: %v", err)
			} else {
				lh.setTransportResponse(snmp, &hostTransport)
			}
		}
		lh.HostTransports[i] = hostTransport

	}
	lh.Write_to_log("DEBUG", "All the ips have been tested")
}

func (lh *LBHost) setTransportResponse(snmpClient *snmplib.SNMP, lbHostTransportResultPayload *LBHostTransportResult) {
	oid, err := snmplib.ParseOid(OID)
	if err != nil {
		lbHostTransportResultPayload.Response_error = fmt.Sprintf("contacted node: Error parsing the OID %v", err)
		return
	}
	pdu, err := snmpClient.GetV3(oid)
	if err != nil {
		lbHostTransportResultPayload.Response_error = fmt.Sprintf("contacted node: The getv3 gave the following error: %v ", err)
		return
	}
	lh.Write_to_log("INFO", fmt.Sprintf("contacted node: transport: %v ip: %v - reply was %v", lbHostTransportResultPayload.Transport, lbHostTransportResultPayload.IP.String(), pdu))
	switch t := pdu.(type) {
	case int:
		lbHostTransportResultPayload.Response_int = pdu.(int)
	case string:
		lbHostTransportResultPayload.Response_string = pdu.(string)
	default:
		lbHostTransportResultPayload.Response_error = fmt.Sprintf("The node returned an unexpected type %s in %v", t, pdu)
	}
}

// todo: instead of polling try adhoc webhook updates
func (lh *LBHost) GetLoadForAlias(clusterName string) int {

	my_load := -200
	for _, my_transport := range lh.HostTransports {
		pduInteger := my_transport.Response_int

		re := regexp.MustCompile(clusterName + "=([0-9]+)")
		submatch := re.FindStringSubmatch(my_transport.Response_string)

		if submatch != nil {
			pduInteger, _ = strconv.Atoi(submatch[1])
		}

		if (pduInteger > 0 && pduInteger < my_load) || (my_load < 0) {
			my_load = pduInteger
		}
		lh.Write_to_log("DEBUG", fmt.Sprintf("Possible load is %v", pduInteger))

	}
	lh.Write_to_log("DEBUG", fmt.Sprintf("THE LOAD IS %v, ", my_load))

	return my_load
}

func (lh *LBHost) GetWorkingIPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range lh.HostTransports {
		if (my_transport.Response_int > 0) && (my_transport.Response_error == "") {
			my_ips = append(my_ips, my_transport.IP)
		}

	}
	lh.Write_to_log("INFO", fmt.Sprintf("The ips for this host are %v", my_ips))
	return my_ips, nil
}

func (lh *LBHost) GetAllIPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range lh.HostTransports {
		my_ips = append(my_ips, my_transport.IP)
	}
	lh.Write_to_log("INFO", fmt.Sprintf("All ips for this host are %v", my_ips))
	return my_ips, nil
}

func (lh *LBHost) GetIps() ([]net.IP, error) {
	var ips []net.IP
	var err error
	re := regexp.MustCompile(".*no such host")
	net.DefaultResolver.StrictErrors = true
	for i := 0; i < 3; i++ {
		lh.Write_to_log("INFO", "Getting the ip addresses")
		ips, err = net.LookupIP(lh.Host_name)
		if err == nil {
			return ips, nil
		}
		lh.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v) ", lh.Host_name, err))
		submatch := re.FindStringSubmatch(err.Error())
		if submatch != nil {
			lh.Write_to_log("INFO", "There is no need to retry this error")
			return nil, err
		}
	}

	lh.Write_to_log("ERROR", "After several retries, we couldn't get the ips!. Let's try with partial results")
	net.DefaultResolver.StrictErrors = false
	ips, err = net.LookupIP(lh.Host_name)
	if err != nil {
		lh.Write_to_log("ERROR", fmt.Sprintf("It didn't work :(. This node will be ignored during this evaluation: %v", err))
	}
	return ips, err
}

func (lh *LBHost) find_transports() {
	lh.Write_to_log("DEBUG", "Let's find the ips behind this host")

	ips, _ := lh.GetIps()
	for _, ip := range ips {
		transport := "udp"
		// If there is an IPv6 address use udp6 transport
		if ip.To4() == nil {
			transport = "udp6"
		}
		lh.HostTransports = append(lh.HostTransports, LBHostTransportResult{Transport: transport,
			Response_int: DefaultResponseInt, Response_string: "", IP: ip,
			Response_error: ""})
	}

}
