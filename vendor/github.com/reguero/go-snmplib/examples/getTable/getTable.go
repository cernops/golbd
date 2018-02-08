package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/deejross/go-snmplib"
)

var target = flag.String("target", "", "The host to connect to")
var community = flag.String("community", "", "The community to use")
var oidString = flag.String("oid", "", "The oid of the table to get")

func doGetTable() {
	flag.Parse()

	fmt.Printf("target=%v\ncommunity=%v\noid=%v\n", *target, *community, *oidString)
	version := snmplib.SNMPv2c

	oid, err := snmplib.ParseOid(*oidString)
	if err != nil {
		fmt.Printf("Error parsing oid '%v' : %v", *oidString, err)
	}

	fmt.Printf("Contacting %v %v %v\n", *target, *community, version)
	snmp, err := snmplib.NewSNMP(*target, *community, version, 2*time.Second, 3)
	if err != nil {
		fmt.Printf("Error creating wsnmp => %v\n", err)
		return
	}
	defer snmp.Close()

	table, err := snmp.GetTable(oid)
	if err != nil {
		fmt.Printf("Error getting table => %v\n", err)
		return
	}
	for k, v := range table {
		fmt.Printf("%v => %v\n", k, v)
	}
}

func main() {
	doGetTable()
}
