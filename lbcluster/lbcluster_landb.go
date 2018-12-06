package lbcluster

import (
	"fmt"
	"github.com/miekg/dns"
	"gitlab.cern.ch/lb-experts/golbd/lbhost"
	"net"
	"sort"
	"strings"
	"time"
)

/* This is the only public function here. It retrieves the status of the dns,
and then updates it with the new hosts */
func (self *LBCluster) Refresh_dns(dnsManager, keyPrefix, internalKey, externalKey string, hosts_to_check map[string]lbhost.LBHost) {

	e := self.get_state_dns(dnsManager)
	if e != nil {
		self.Write_to_log("WARNING", fmt.Sprintf("Get_state_dns Error: %v", e.Error()))
	}
	e = self.update_dns(keyPrefix+"internal.", internalKey, dnsManager, hosts_to_check)
	if e != nil {
		self.Write_to_log("WARNING", fmt.Sprintf("Internal Update_dns Error: %v", e.Error()))
	}
	if self.externally_visible() {
		e = self.update_dns(keyPrefix+"external.", externalKey, dnsManager, hosts_to_check)
		if e != nil {
			self.Write_to_log("WARNING", fmt.Sprintf("External Update_dns Error: %v", e.Error()))
		}
	}
}

//Internal functions
func (self *LBCluster) externally_visible() bool {
	return self.Parameters.External
}

func (self *LBCluster) update_dns(keyName, tsigKey, dnsManager string, hosts_to_check map[string]lbhost.LBHost) error {
	pbhDns := strings.Join(self.Previous_best_hosts_dns, " ")
	cbh := strings.Join(self.Current_best_hosts, " ")
	if pbhDns == cbh {
		self.Write_to_log("INFO", fmt.Sprintf("DNS not update keyName %v cbh == pbhDns == %v", keyName, cbh))
		return nil
	}
	cluster_name := self.Cluster_name
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	ttl := "60"
	if self.Parameters.Ttl > 60 {
		ttl = fmt.Sprintf("%d", self.Parameters.Ttl)
	}
	//best_hosts_len := len(self.Current_best_hosts)
	m := new(dns.Msg)
	m.SetUpdate(cluster_name + ".")
	m.Id = 1234
	rr_removeA, _ := dns.NewRR(cluster_name + ". " + ttl + " IN A 127.0.0.1")
	rr_removeAAAA, _ := dns.NewRR(cluster_name + ". " + ttl + " IN AAAA ::1")
	m.RemoveRRset([]dns.RR{rr_removeA})
	m.RemoveRRset([]dns.RR{rr_removeAAAA})

	for _, hostname := range self.Current_best_hosts {
		my_host := hosts_to_check[hostname]
		ips, err := my_host.Get_working_IPs()
		if err != nil {
			self.Write_to_log("WARNING", fmt.Sprintf("LookupIP: %v has incorrect or missing IP address (%v)", hostname, err))
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
	self.Write_to_log("INFO", fmt.Sprintf("WE WOULD UPDATE THE DNS WITH THE IPS %v", m))
	c := new(dns.Client)
	m.SetTsig(keyName, dns.HmacMD5, 300, time.Now().Unix())
	c.TsigSecret = map[string]string{keyName: tsigKey}
	_, _, err := c.Exchange(m, dnsManager+":53")
	if err != nil {
		self.Write_to_log("ERROR", fmt.Sprintf("DNS update failed with (%v)", err))
	}
	self.Write_to_log("INFO", fmt.Sprintf("DNS update with keyName %v", keyName))
	return err
}

func (self *LBCluster) get_state_dns(dnsManager string) error {
	cluster_name := self.Cluster_name
	if !strings.HasSuffix(cluster_name, ".cern.ch") {
		cluster_name = cluster_name + ".cern.ch"
	}
	m := new(dns.Msg)
	m.SetEdns0(4096, false)
	m.SetQuestion(cluster_name+".", dns.TypeA)
	in, err := dns.Exchange(m, dnsManager+":53")
	if err != nil {
		self.Write_to_log("ERROR", fmt.Sprintf("Error getting the ipv4 state of dns: %v", err))
		return err
	}

	var ips []net.IP
	for _, a := range in.Answer {
		if t, ok := a.(*dns.A); ok {
			self.Slog.Debug(fmt.Sprintf("%v", t))
			self.Slog.Debug(fmt.Sprintf("%v", t.A))
			ips = append(ips, t.A)
		}
	}
	m.SetQuestion(cluster_name+".", dns.TypeAAAA)
	in, err = dns.Exchange(m, dnsManager+":53")
	if err != nil {
		self.Write_to_log("ERROR", fmt.Sprintf("Error getting the ipv6 state of dns: %v", err))
		return err
	}

	for _, a := range in.Answer {
		if t, ok := a.(*dns.AAAA); ok {
			self.Slog.Debug(fmt.Sprintf("%v", t))
			self.Slog.Debug(fmt.Sprintf("%v", t.AAAA))
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
			self.Write_to_log("ERROR", fmt.Sprintf("Error getting the state of the dns %v", err))
			return err
		}
		if len(names) > 0 {
			if len(names) == 1 {
				name = strings.TrimRight(names[0], ".")
			} else {
				name, err = net.LookupCNAME(names[0])
				if err != nil {
					self.Write_to_log("ERROR", fmt.Sprintf("Error getting the state of the dns %v", err))
					return err
				}
				name = strings.TrimRight(name, ".")
			}
			host_list = append(host_list, name)
		}
	}
	removeDuplicates(&host_list)
	self.Previous_best_hosts_dns = host_list
	prevBesthostsDns := make([]string, len(self.Previous_best_hosts_dns))
	prevBesthosts := make([]string, len(self.Previous_best_hosts))
	currBesthosts := make([]string, len(self.Current_best_hosts))
	copy(prevBesthostsDns, self.Previous_best_hosts_dns)
	copy(prevBesthosts, self.Previous_best_hosts)
	copy(currBesthosts, self.Current_best_hosts)
	sort.Strings(prevBesthostsDns)
	sort.Strings(prevBesthosts)
	sort.Strings(currBesthosts)
	pbhDns := strings.Join(prevBesthostsDns, " ")
	pbh := strings.Join(prevBesthosts, " ")
	cbh := strings.Join(currBesthosts, " ")
	if pbh != "unknown" {
		if pbh != pbhDns {
			self.Write_to_log("WARNING", "Prev DNS state "+pbhDns+" - Prev local state  "+pbh+" differ")
		}
	}
	if cbh == "unknown" {
		self.Write_to_log("WARNING", "Current best hosts are unknown - Taking Previous DNS state  "+pbhDns)
		self.Current_best_hosts = self.Previous_best_hosts_dns
	}
	return err
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
