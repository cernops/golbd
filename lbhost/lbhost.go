package lbhost

import (
	"fmt"
	"lb-experts/golbd/logger"
	"lb-experts/golbd/model"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/reguero/go-snmplib"
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
	ClusterConfig  model.ClusterConfig
	Host_name      string
	HostTransports []LBHostTransportResult
	Logger         logger.Logger
	SnmpAgent      DiscoveryAgent
}

type snmpDiscoveryResult struct {
	hostIdx             int
	hostTransportResult LBHostTransportResult
}

type DiscoveryAgent interface {
	Close() error
	Discover() error
	GetV3(oid snmplib.Oid) (interface{}, error)
}

func NewHostDiscoveryAgent(nodeIp string, clusterConfig model.ClusterConfig) (DiscoveryAgent, error) {
	return snmplib.NewSNMPv3(nodeIp, clusterConfig.Loadbalancing_username, "MD5",
		clusterConfig.Loadbalancing_password, "NOPRIV", clusterConfig.Loadbalancing_password,
		time.Duration(TIMEOUT)*time.Second, 2)
}

type Host interface {
	GetName() string
	SetName(name string)
	SNMPDiscovery()
	GetClusterConfig() *model.ClusterConfig
	GetLoadForAlias(clusterName string) int
	GetWorkingIPs() ([]net.IP, error)
	GetAllIPs() ([]net.IP, error)
	GetIps() ([]net.IP, error)
	SetTransportPayload(transportPayloadList []LBHostTransportResult)
	GetHostTransportPayloads() []LBHostTransportResult
}

func NewLBHost(clusterConfig model.ClusterConfig, logger logger.Logger) Host {
	return &LBHost{
		ClusterConfig: clusterConfig,
		Logger:        logger,
	}
}

func (lh *LBHost) SetName(name string) {
	lh.Host_name = name
}

func (lh *LBHost) GetName() string {
	return lh.Host_name
}
func (lh *LBHost) GetClusterConfig() *model.ClusterConfig {
	return &lh.ClusterConfig
}

func (lh *LBHost) GetHostTransportPayloads() []LBHostTransportResult {
	return lh.HostTransports
}

func (lh *LBHost) SetTransportPayload(transportPayloadList []LBHostTransportResult) {
	lh.HostTransports = transportPayloadList
}

func (lh *LBHost) SNMPDiscovery() {
	var wg sync.WaitGroup
	lh.find_transports()
	discoveryResultChan := make(chan snmpDiscoveryResult)
	defer close(discoveryResultChan)
	hostTransportResultList := make([]LBHostTransportResult, 0, len(lh.HostTransports))
	hostTransportResultList = append(hostTransportResultList, lh.HostTransports...)
	for i, hostTransport := range lh.HostTransports {
		wg.Add(1)
		go func(idx int, hostTransport LBHostTransportResult) {
			defer wg.Done()
			lh.discoverNode(idx, hostTransport, discoveryResultChan)
		}(i, hostTransport)
	}
	go func(discoveryResultChan <-chan snmpDiscoveryResult) {
		for discoveryResultData := range discoveryResultChan {
			hostTransportResultList[discoveryResultData.hostIdx] = discoveryResultData.hostTransportResult
		}
	}(discoveryResultChan)
	wg.Wait()
	lh.HostTransports = hostTransportResultList
	lh.Logger.Debug("All the ips have been tested")
}

func (lh *LBHost) discoverNode(hostTransportIdx int, hostTransport LBHostTransportResult, resultChan chan<- snmpDiscoveryResult) {
	var snmpAgent DiscoveryAgent
	var err error
	hostTransport.Response_int = DefaultResponseInt
	nodeIp := hostTransport.IP.String()
	lh.Logger.Debug("Checking the host " + nodeIp + " with " + hostTransport.Transport)
	if lh.SnmpAgent == nil {
		snmpAgent, err = NewHostDiscoveryAgent(nodeIp, lh.ClusterConfig)
		if err != nil {
			hostTransport.Response_error = fmt.Sprintf("contacted node: error creating the snmp object: %v", err)
		}
	} else {
		snmpAgent = lh.SnmpAgent
	}
	if err == nil {
		defer snmpAgent.Close()
		err = snmpAgent.Discover()
		if err != nil {
			hostTransport.Response_error = fmt.Sprintf("contacted node: error in the snmp discovery: %v", err)
		} else {
			lh.setTransportResponse(snmpAgent, &hostTransport)
		}
	}

	resultChan <- snmpDiscoveryResult{
		hostIdx:             hostTransportIdx,
		hostTransportResult: hostTransport,
	}
}

func (lh *LBHost) setTransportResponse(snmpClient DiscoveryAgent, lbHostTransportResultPayload *LBHostTransportResult) {
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
	lh.Logger.Info(fmt.Sprintf("contacted node: transport: %v ip: %v - reply was %v", lbHostTransportResultPayload.Transport, lbHostTransportResultPayload.IP.String(), pdu))
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
		lh.Logger.Debug(fmt.Sprintf("Possible load is %v", pduInteger))

	}
	lh.Logger.Debug(fmt.Sprintf("THE LOAD IS %v, ", my_load))

	return my_load
}

func (lh *LBHost) GetWorkingIPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range lh.HostTransports {
		if (my_transport.Response_int > 0) && (my_transport.Response_error == "") {
			my_ips = append(my_ips, my_transport.IP)
		}

	}
	lh.Logger.Info(fmt.Sprintf("The ips for this host are %v", my_ips))
	return my_ips, nil
}

func (lh *LBHost) GetAllIPs() ([]net.IP, error) {
	var my_ips []net.IP
	for _, my_transport := range lh.HostTransports {
		my_ips = append(my_ips, my_transport.IP)
	}
	lh.Logger.Info(fmt.Sprintf("All ips for this host are %v", my_ips))
	return my_ips, nil
}

func (lh *LBHost) GetIps() ([]net.IP, error) {
	var ips []net.IP
	var err error
	re := regexp.MustCompile(".*no such host")
	net.DefaultResolver.StrictErrors = true
	for i := 0; i < 3; i++ {
		lh.Logger.Info("Getting the ip addresses")
		ips, err = net.LookupIP(lh.Host_name)
		if err == nil {
			return ips, nil
		}
		lh.Logger.Info(fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v) ", lh.Host_name, err))
		submatch := re.FindStringSubmatch(err.Error())
		if submatch != nil {
			lh.Logger.Info("There is no need to retry this error")
			return nil, err
		}
	}

	lh.Logger.Error("After several retries, we couldn't get the ips!. Let's try with partial results")
	net.DefaultResolver.StrictErrors = false
	ips, err = net.LookupIP(lh.Host_name)
	if err != nil {
		lh.Logger.Error(fmt.Sprintf("It didn't work :(. This node will be ignored during this evaluation: %v", err))
	}
	return ips, err
}

func (lh *LBHost) find_transports() {
	lh.Logger.Debug("Let's find the ips behind this host")

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
