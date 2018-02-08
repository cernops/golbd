package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/deejross/go-snmplib"
)

func myUDPServer(listenIPAddr string, port int) *net.UDPConn {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(listenIPAddr),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("udp Listen error.")
		panic(err)
	}
	return conn
}

func main() {
	rand.Seed(0)
	target := ""
	community := ""
	version := snmplib.SNMPv2c

	udpsock := myUDPServer("0.0.0.0", 162)
	defer udpsock.Close()

	snmp := snmplib.NewSNMPOnConn(target, community, version, 2*time.Second, 5, udpsock)
	defer snmp.Close()

	snmp.TrapUsers = append(snmp.TrapUsers, snmplib.V3user{"pcb.snmpv3", "SHA1", "this_is_my_pcb", "AES", "my_pcb_is_4_me"})

	packet := make([]byte, 3000)
	for {
		_, addr, err := udpsock.ReadFromUDP(packet)
		if err != nil {
			log.Fatal("udp read error\n")
		}

		log.Printf("Received trap from %s:\n", addr.IP)

		varbinds, err := snmp.ParseTrap(packet)
		if err != nil {
			log.Printf("Error processing trap: %v.", err)
			continue
		}

		fmt.Println(varbinds)
	}
}
