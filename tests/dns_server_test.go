package main_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// parseQuery handles the basic query of RRs
func parseQuery(m *dns.Msg, records map[string][]string) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			if ips, ok := records[q.Name]; ok {
				for _, ip := range ips {
					if strings.Contains(ip, ":") {
						continue
					}
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		case dns.TypeAAAA:
			if ips, ok := records[q.Name]; ok {
				for _, ip := range ips {
					if !strings.Contains(ip, ":") {
						continue
					}
					rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}
}

// parseUpdate handles the dynamic update of RRs
func parseUpdate(r *dns.Msg, records map[string][]string) {
	for _, question := range r.Question {
		for _, rr := range r.Ns {
			if rr == nil {
				// Delete
				delete(records, question.Name)
			} else {
				// Add
				if a, ok := rr.(*dns.A); ok {
					records[question.Name] = append(records[question.Name], a.A.String())
				} else if aaaa, ok := rr.(*dns.AAAA); ok {
					records[question.Name] = append(records[question.Name], aaaa.AAAA.String())
				}
			}
		}
	}
}

// handleDnsRequest delegate the dns request to the approriate parser
func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg, records map[string][]string) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m, records)
	case dns.OpcodeUpdate:
		parseUpdate(r, records)
	}

	m.Rcode = 0
	m.Authoritative = true
	w.WriteMsg(m)
}

// setupDNSServer creates a simple DNS server and listens on the port specified
// Adapted from Andreas Wålm's Gist https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb
func setupDnsServer(port string) (*dns.Server, error) {
	var records = map[string][]string{
		"aiermis.cern.ch.":    {"188.184.104.111", "2001:1458:d00:2d::100:58"},
		"testrefresh.cern.ch": {"1.2.3.4"},
		"nochange.cern.ch":    {"1.1.1.1"},
	}

	// Create a local dns server
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) { handleDnsRequest(w, r, records) })

	dnsServerStarted := make(chan bool)
	notifyStartedFunc := func() {
		dnsServerStarted <- true
	}

	server := &dns.Server{Addr: ":" + port, Net: "udp", NotifyStartedFunc: notifyStartedFunc}
	go server.ListenAndServe()

	// Wait for the DNS server to start
	timeOut := time.After(2 * time.Second)
	for {
		select {
		case <-dnsServerStarted:
			return server, nil
		case <-timeOut:
			defer server.Shutdown()
			return nil, errors.New("DNS server does not start within the time limit")
		}
	}
}
