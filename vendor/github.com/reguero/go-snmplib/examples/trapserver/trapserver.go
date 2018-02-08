// Package main provides an example of how to create an SNMP trap receiver server.
// NOTE: In Unix environments (Mac, Linux, etc) this needs to be run as root or
// using "sudo" since this requires port 162 to run and ports in those environments
// below 1024 are protected. You may get a "permission denied" error trying to run
// this without root or "sudo".
package main

import (
	"encoding/json"
	"log"
	"net"

	snmplib "github.com/deejross/go-snmplib"
)

type snmpHandler struct{}

func (h snmpHandler) OnError(addr net.Addr, err error) {
	log.Println(addr.String(), err)
}

func (h snmpHandler) OnTrap(addr net.Addr, trap snmplib.Trap) {
	prettyPrint, _ := json.MarshalIndent(trap, "", "\t")
	log.Println(string(prettyPrint))
}

func main() {
	server, err := snmplib.NewTrapServer("0.0.0.0", 162)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Listening for traps on port 162")
	server.ListenAndServe(snmpHandler{})
}
