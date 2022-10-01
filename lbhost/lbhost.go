package lbhost

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reguero/go-snmplib"
)

const (
	TIMEOUT int    = 10
	OID     string = ".1.3.6.1.4.1.96.255.1"
)

type TransportResult struct {
	Transport      string
	IP             net.IP
	ResponseInt    int
	ResponseString string
	ResponseError  string
}
type LBHost struct {
	ClusterName    string
	HostName       string
	HostTransports []TransportResult
	LBUsername     string
	LBPassword     string
	LogFile        string
	logMu          sync.Mutex
	DebugFlag      bool
}

func (h *LBHost) SnmpReq() {

	h.findTransports()

	for i, myTransport := range h.HostTransports {
		myTransport.ResponseInt = 100000
		transport := myTransport.Transport
		nodeIp := myTransport.IP.String()

		/* There is no need to put square brackets around the ipv6 addresses*/
		_ = h.WriteToLog("DEBUG", "Checking the host "+nodeIp+" with "+transport)
		snmp, err := snmplib.NewSNMPv3(nodeIp, h.LBUsername, "MD5", h.LBPassword, "NOPRIV", h.LBPassword,
			time.Duration(TIMEOUT)*time.Second, 2)
		if err != nil {
			// Failed to create snmpgo.SNMP object
			myTransport.ResponseError = fmt.Sprintf("contacted node: error creating the snmp object: %v", err)
		} else {
			defer snmp.Close()
			err = snmp.Discover()

			if err != nil {
				myTransport.ResponseError = fmt.Sprintf("contacted node: error in the snmp discovery: %v", err)

			} else {

				oid, err := snmplib.ParseOid(OID)

				if err != nil {
					// Failed to parse Oids
					myTransport.ResponseError = fmt.Sprintf("contacted node: Error parsing the OID %v", err)

				} else {
					pdu, err := snmp.GetV3(oid)

					if err != nil {
						myTransport.ResponseError = fmt.Sprintf("contacted node: The getv3 gave the following error: %v ", err)

					} else {

						_ = h.WriteToLog("INFO", fmt.Sprintf("contacted node: transport: %v ip: %v - reply was %v", transport, nodeIp, pdu))

						//var pduInteger int
						switch t := pdu.(type) {
						case int:
							myTransport.ResponseInt = pdu.(int)
						case string:
							myTransport.ResponseString = pdu.(string)
						default:
							myTransport.ResponseError = fmt.Sprintf("The node returned an unexpected type %s in %v", t, pdu)
						}
					}
				}
			}
		}
		h.HostTransports[i] = myTransport

	}

	_ = h.WriteToLog("DEBUG", "All the ips have been tested")
}

func (h *LBHost) WriteToLog(level string, msg string) error {
	var err error
	if level == "DEBUG" && !h.DebugFlag {
		//The debug messages should not appear
		return nil
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	timestamp := time.Now().Format(time.StampMilli)
	msg = fmt.Sprintf("%s lbd[%d]: %s: cluster: %s node: %s %s", timestamp, os.Getpid(), level, h.ClusterName, h.HostName, msg)

	h.logMu.Lock()
	defer h.logMu.Unlock()

	f, err := os.OpenFile(h.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, msg)

	return err
}

func (h *LBHost) GetLoadForAlias(clusterName string) int {

	load := -200
	for _, transport := range h.HostTransports {
		pduInteger := transport.ResponseInt

		re := regexp.MustCompile(clusterName + "=([0-9]+)")
		submatch := re.FindStringSubmatch(transport.ResponseString)

		if submatch != nil {
			pduInteger, _ = strconv.Atoi(submatch[1])
		}

		if (pduInteger > 0 && pduInteger < load) || (load < 0) {
			load = pduInteger
		}
		_ = h.WriteToLog("DEBUG", fmt.Sprintf("Possible load is %v", pduInteger))

	}
	_ = h.WriteToLog("DEBUG", fmt.Sprintf("THE LOAD IS %v, ", load))

	return load
}

func (h *LBHost) GetWorkingIPs() ([]net.IP, error) {
	var ips []net.IP
	for _, transport := range h.HostTransports {
		if (transport.ResponseInt > 0) && (transport.ResponseError == "") {
			ips = append(ips, transport.IP)
		}

	}
	_ = h.WriteToLog("INFO", fmt.Sprintf("The ips for this host are %v", ips))
	return ips, nil
}

func (h *LBHost) GetAllIPs() ([]net.IP, error) {
	var ips []net.IP
	for _, transport := range h.HostTransports {
		ips = append(ips, transport.IP)
	}
	_ = h.WriteToLog("INFO", fmt.Sprintf("All ips for this host are %v", ips))
	return ips, nil
}

func (h *LBHost) GetIps() ([]net.IP, error) {

	var ips []net.IP

	var err error

	re := regexp.MustCompile(".*no such host")

	net.DefaultResolver.StrictErrors = true

	for i := 0; i < 3; i++ {
		_ = h.WriteToLog("INFO", "Getting the ip addresses")
		ips, err = net.LookupIP(h.HostName)
		if err == nil {
			return ips, nil
		}
		_ = h.WriteToLog("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v) ", h.HostName, err))
		submatch := re.FindStringSubmatch(err.Error())
		if submatch != nil {
			_ = h.WriteToLog("INFO", "There is no need to retry this error")
			return nil, err
		}
	}

	_ = h.WriteToLog("ERROR", "After several retries, we couldn't get the ips!. Let's try with partial results")
	net.DefaultResolver.StrictErrors = false
	ips, err = net.LookupIP(h.HostName)
	if err != nil {
		_ = h.WriteToLog("ERROR", fmt.Sprintf("It didn't work :(. This node will be ignored during this evaluation: %v", err))
	}
	return ips, err
}

func (h *LBHost) findTransports() {
	_ = h.WriteToLog("DEBUG", "Let's find the ips behind this host")

	ips, _ := h.GetIps()
	for _, ip := range ips {
		transport := "udp"
		// If there is an IPv6 address use udp6 transport
		if ip.To4() == nil {
			transport = "udp6"
		}
		h.HostTransports = append(h.HostTransports, TransportResult{Transport: transport,
			ResponseInt: 100000, ResponseString: "", IP: ip,
			ResponseError: ""})
	}
}
