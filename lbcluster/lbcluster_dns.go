package lbcluster

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

/*RefreshDNS This is the only public function here. It retrieves the current ips behind the dns,
and then updates it with the new best ips (if they are different) */
func (lbc *LBCluster) RefreshDNS(dnsManager, keyPrefix, internalKey, externalKey string) {

	e := lbc.GetStateDNS(dnsManager)
	if e != nil {
		lbc.Write_to_log("WARNING", fmt.Sprintf("Get_state_dns Error: %v", e.Error()))
	}

	pbiDNS := lbc.concatenateIps(lbc.Previous_best_ips_dns)
	cbi := lbc.concatenateIps(lbc.Current_best_ips)
	if pbiDNS == cbi {
		lbc.Write_to_log("INFO", fmt.Sprintf("DNS not update keyName %v cbh == pbhDns == %v", keyPrefix, cbi))
		return
	}

	lbc.Write_to_log("INFO", fmt.Sprintf("Updating the DNS with %v (previous state was %v)", cbi, pbiDNS))

	e = lbc.updateDNS(keyPrefix+"internal.", internalKey, dnsManager)
	if e != nil {
		lbc.Write_to_log("WARNING", fmt.Sprintf("Internal Update_dns Error: %v", e.Error()))
	}
	if lbc.externallyVisible() {
		e = lbc.updateDNS(keyPrefix+"external.", externalKey, dnsManager)
		if e != nil {
			lbc.Write_to_log("WARNING", fmt.Sprintf("External Update_dns Error: %v", e.Error()))
		}
	}
}

//Internal functions
func (lbc *LBCluster) externallyVisible() bool {
	return lbc.Parameters.External
}

func (lbc *LBCluster) updateDNS(keyName, tsigKey, dnsManager string) error {

	ttl := "60"
	if lbc.Parameters.Ttl > 60 {
		ttl = fmt.Sprintf("%d", lbc.Parameters.Ttl)
	}
	//best_hosts_len := len(lbc.Current_best_hosts)
	m := new(dns.Msg)
	m.SetUpdate(lbc.Cluster_name + ".")
	m.Id = 1234
	rrRemoveA, _ := dns.NewRR(lbc.Cluster_name + ". " + ttl + " IN A 127.0.0.1")
	rrRemoveAAAA, _ := dns.NewRR(lbc.Cluster_name + ". " + ttl + " IN AAAA ::1")
	m.RemoveRRset([]dns.RR{rrRemoveA})
	m.RemoveRRset([]dns.RR{rrRemoveAAAA})

	for _, ip := range lbc.Current_best_ips {
		var rrInsert dns.RR
		if ip.To4() != nil {
			rrInsert, _ = dns.NewRR(lbc.Cluster_name + ". " + ttl + " IN A " + ip.String())
		} else if ip.To16() != nil {
			rrInsert, _ = dns.NewRR(lbc.Cluster_name + ". " + ttl + " IN AAAA " + ip.String())
		}
		m.Insert([]dns.RR{rrInsert})
	}
	lbc.Write_to_log("INFO", fmt.Sprintf("WE WOULD UPDATE THE DNS WITH THE IPS %v", m))
	c := new(dns.Client)
	m.SetTsig(keyName, dns.HmacMD5, 300, time.Now().Unix())
	c.TsigSecret = map[string]string{keyName: tsigKey}
	_, _, err := c.Exchange(m, dnsManager)
	if err != nil {
		lbc.Write_to_log("ERROR", fmt.Sprintf("DNS update failed with (%v)", err))
		return err
	}
	lbc.Write_to_log("INFO", fmt.Sprintf("DNS update with keyName %v", keyName))

	return nil
}

func (lbc *LBCluster) getIpsFromDNS(m *dns.Msg, dnsManager string, dnsType uint16, ips *[]net.IP) error {
	m.SetQuestion(lbc.Cluster_name+".", dnsType)
	in, err := dns.Exchange(m, dnsManager)
	if err != nil {
		lbc.Write_to_log("ERROR", fmt.Sprintf("Error getting the ipv4 state of dns: %v", err))
		return err
	}
	for _, a := range in.Answer {
		if t, ok := a.(*dns.A); ok {
			lbc.Slog.Debug(fmt.Sprintf("From %v, got ipv4 %v", t, t.A))
			*ips = append(*ips, t.A)
		} else if t, ok := a.(*dns.AAAA); ok {
			lbc.Slog.Debug(fmt.Sprintf("From %v, got ipv6 %v", t, t.AAAA))
			*ips = append(*ips, t.AAAA)
		}
	}
	return nil
}

func (lbc *LBCluster) GetStateDNS(dnsManager string) error {
	m := new(dns.Msg)
	var ips []net.IP
	m.SetEdns0(4096, false)
	lbc.Write_to_log("DEBUG", "Getting the ips from the DNS")
	err := lbc.getIpsFromDNS(m, dnsManager, dns.TypeA, &ips)

	if err != nil {
		return err
	}
	err = lbc.getIpsFromDNS(m, dnsManager, dns.TypeAAAA, &ips)
	if err != nil {
		return err
	}

	lbc.Write_to_log("INFO", fmt.Sprintf("Let's keep the list of ips : %v", ips))
	lbc.Previous_best_ips_dns = ips

	return nil
}
