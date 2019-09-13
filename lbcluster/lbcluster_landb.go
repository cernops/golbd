package lbcluster

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
)

/* This is the only public function here. It retrieves the status of the dns,
and then updates it with the new hosts */
func (lbc *LBCluster) Refresh_dns(dnsManager, keyPrefix, internalKey, externalKey string, hosts_to_check map[string]lbhost.LBHost) {

	_, e := lbc.get_state_dns(dnsManager)
	if e != nil {
		lbc.Write_to_log("WARNING", fmt.Sprintf("Get_state_dns Error: %v", e.Error()))
		return
	}
	e = lbc.update_dns(keyPrefix+"internal.", internalKey, dnsManager, hosts_to_check)
	if e != nil {
		lbc.Write_to_log("WARNING", fmt.Sprintf("Internal Update_dns Error: %v", e.Error()))
	}
	if lbc.externallyVisible() {
		e = lbc.update_dns(keyPrefix+"external.", externalKey, dnsManager, hosts_to_check)
		if e != nil {
			lbc.Write_to_log("WARNING", fmt.Sprintf("External Update_dns Error: %v", e.Error()))
		}
	}
}

//Internal functions
func (lbc *LBCluster) externallyVisible() bool {
	return lbc.Parameters.External
}

func (lbc *LBCluster) update_dns(keyName, tsigKey, dnsManager string, hosts_to_check map[string]lbhost.LBHost) error {
	pbhDns := strings.Join(lbc.Previous_best_hosts_dns, " ")
	cbh := strings.Join(lbc.Current_best_hosts, " ")
	if pbhDns == cbh {
		lbc.Write_to_log("INFO", fmt.Sprintf("DNS not update keyName %v cbh == pbhDns == %v", keyName, cbh))
		return nil
	}
	lbc.Write_to_log("INFO", fmt.Sprintf("Updating the DNS with %v (previous state was %v)", cbh, pbhDns))
	cluster_name := lbc.Cluster_name
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	ttl := "60"
	if lbc.Parameters.Ttl > 60 {
		ttl = fmt.Sprintf("%d", lbc.Parameters.Ttl)
	}
	//best_hosts_len := len(lbc.Current_best_hosts)
	m := new(dns.Msg)
	m.SetUpdate(cluster_name + ".")
	m.Id = 1234
	rr_removeA, _ := dns.NewRR(cluster_name + ". " + ttl + " IN A 127.0.0.1")
	rr_removeAAAA, _ := dns.NewRR(cluster_name + ". " + ttl + " IN AAAA ::1")
	m.RemoveRRset([]dns.RR{rr_removeA})
	m.RemoveRRset([]dns.RR{rr_removeAAAA})

	for _, hostname := range lbc.Current_best_hosts {
		my_host := hosts_to_check[hostname]
		ips, err := my_host.Get_working_IPs()
		if err != nil {
			lbc.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", hostname, err))
			continue
		}
		for _, ip := range ips {
			var rr_insert dns.RR
			if ip.To4() != nil {
				rr_insert, _ = dns.NewRR(cluster_name + ". " + ttl + " IN A " + ip.String())
			} else if ip.To16() != nil {
				rr_insert, _ = dns.NewRR(cluster_name + ". " + ttl + " IN AAAA " + ip.String())
			}
			m.Insert([]dns.RR{rr_insert})
		}
	}
	lbc.Write_to_log("INFO", fmt.Sprintf("WE WOULD UPDATE THE DNS WITH THE IPS %v", m))
	c := new(dns.Client)
	m.SetTsig(keyName, dns.HmacMD5, 300, time.Now().Unix())
	c.TsigSecret = map[string]string{keyName: tsigKey}
	_, _, err := c.Exchange(m, dnsManager+":53")
	if err != nil {
		lbc.Write_to_log("ERROR", fmt.Sprintf("DNS update failed with (%v)", err))
	}
	lbc.Write_to_log("INFO", fmt.Sprintf("DNS update with keyName %v", keyName))
	return err
}

func (lbc *LBCluster) get_state_dns(dnsManager string) ([]net.IP, error) {
	cluster_name := lbc.Cluster_name
	var ips []net.IP
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	m := new(dns.Msg)
	m.SetEdns0(4096, false)
	m.SetQuestion(cluster_name+".", dns.TypeA)
	in, err := dns.Exchange(m, dnsManager+":53")
	if err != nil {
		lbc.Write_to_log("ERROR", fmt.Sprintf("Error getting the ipv4 state of dns: %v", err))
		return nil, err
	}
	//fmt.Println(in)
	for _, a := range in.Answer {
		if t, ok := a.(*dns.A); ok {
			lbc.Slog.Debug(fmt.Sprintf("%v", t))
			lbc.Slog.Debug(fmt.Sprintf("%v", t.A))
			ips = append(ips, t.A)
		}
	}
	m.SetQuestion(cluster_name+".", dns.TypeAAAA)
	in, err = dns.Exchange(m, dnsManager+":53")
	if err != nil {
		lbc.Write_to_log("ERROR", fmt.Sprintf("Error getting the ipv6 state of dns: %v", err))
		return ips, err
	}

	for _, a := range in.Answer {
		if t, ok := a.(*dns.AAAA); ok {
			lbc.Slog.Debug(fmt.Sprintf("%v", t))
			lbc.Slog.Debug(fmt.Sprintf("%v", t.AAAA))
			ips = append(ips, t.AAAA)
		}
	}
	// Check if there is any host behind the alias
	if len(ips) == 0 {
		return ips, nil
	}

	//addrs, err := net.LookupHost(cluster_name)
	//ips, err := net.LookupIP(cluster_name)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(ips)
	var name string
	var host_list []string

	net.DefaultResolver.StrictErrors = true

	for _, ip := range ips {
		names, err := net.LookupAddr(ip.String())
		if err != nil {
			lbc.Write_to_log("ERROR", fmt.Sprintf("Error getting the state of the dns %v", err))
			if _, ok := err.(*net.DNSError); ok {
				lbc.Write_to_log("INFO", "The host does not exist anymore... let's continue")
				err = nil
			} else {
				lbc.Write_to_log("INFO", "Different error")
				return ips, err
			}
		}

		if len(names) > 0 {
			if len(names) == 1 {
				name = strings.TrimRight(names[0], ".")
			} else {
				name, err = net.LookupCNAME(names[0])
				if err != nil {
					lbc.Write_to_log("ERROR", fmt.Sprintf("Error getting the state of the dns %v", err))
					return ips, err
				}
				name = strings.TrimRight(name, ".")
			}
			host_list = append(host_list, name)
		}
	}
	net.DefaultResolver.StrictErrors = false

	removeDuplicates(&host_list)
	// The order in which the DNS returns the hosts does not seem to depend on the order in which
	// they have been set. Let's sort them to avoid unnecessary udpates
	sort.Strings(host_list)
	lbc.Previous_best_hosts_dns = host_list

	pbhDns := strings.Join(lbc.Previous_best_hosts_dns, " ")
	pbh := strings.Join(lbc.Previous_best_hosts, " ")
	cbh := strings.Join(lbc.Current_best_hosts, " ")
	if pbh != "unknown" {
		if pbh != pbhDns {
			lbc.Write_to_log("WARNING", "Prev DNS state "+pbhDns+" - Prev local state  "+pbh+" differ")
		}
	}
	if cbh == "unknown" {
		lbc.Write_to_log("WARNING", "Current best hosts are unknown - Taking Previous DNS state  "+pbhDns)
		lbc.Current_best_hosts = lbc.Previous_best_hosts_dns
	}
	return ips, err
}

// Internal functions
//
func removeDuplicates(xs *[]string) {
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
