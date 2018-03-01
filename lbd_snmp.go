package main

import (
	"fmt"
	"github.com/reguero/go-snmplib"
	"time"
)

func exampleMain() {
	fmt.Printf("HELLO WORLD")
	oidstr := ".1.3.6.1.4.1.96.255.1"
	oid := snmplib.MustParseOid(oidstr)
	fmt.Printf("Got the oid %v", oid)

	target := "ermis12.cern.ch"
	username := "loadbalancing"
	authAlg := "MD5"
	authKey := "XXXXXXXXXXXXXXX"
	privAlg := "NOPRIV" //"DES"
	privKey := authKey

	wsnmp, err := snmplib.NewSNMPv3(target, username, authAlg, authKey, privAlg, privKey, 2*time.Second, 2)
	if err != nil {
		fmt.Printf("Error creating wsnmp => %v\n%v", wsnmp, err)

		return
	}
	defer wsnmp.Close()
	err = wsnmp.Discover()
	if err != nil {
		fmt.Printf("Error in the discover")
		return
	}
	fmt.Printf("Ready to do the get\n")

	val, err := wsnmp.GetV3(oid)

	fmt.Printf("Got the val %v", val)
	fmt.Printf("And the error %v", err)

	fmt.Printf("CALL DONE")
}
