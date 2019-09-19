package main_test

import (
	"fmt"
	"log"
	"time"

	snmplib "github.com/reguero/go-snmplib"
)

func Example() {
	fmt.Printf("HELLO WORLD\n")
	oidstr := ".1.3.6.1.4.1.96.255.1"
	oid := snmplib.MustParseOid(oidstr)
	fmt.Printf("Got the oid %v\n", oid)

	target := "ermis12.cern.ch"
	username := "loadbalancing"
	authAlg := "MD5"
	authKey := "XXXXXXXXXXXXXXX"
	privAlg := "NOPRIV" //"DES"
	privKey := authKey

	wsnmp, err := snmplib.NewSNMPv3(target, username, authAlg, authKey, privAlg, privKey, 2*time.Second, 2)
	if err != nil {
		log.Fatalf("Error creating wsnmp => %v\n%v\n", wsnmp, err)
	}
	defer wsnmp.Close()

	err = wsnmp.Discover()
	if err != nil {
		log.Fatalf("Could not discover: %v", err)
		return
	}
	fmt.Printf("Ready to do the get\n")

	val, err := wsnmp.GetV3(oid)

	fmt.Printf("Got the val %v\n", val)
	fmt.Printf("And the error %v\n", err)

	fmt.Printf("CALL DONE\n")
}
