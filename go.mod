module lb-experts/golbd

go 1.17

require (
	github.com/miekg/dns v1.0.0
	github.com/reguero/go-snmplib v0.0.0-20181019092238-e566f5619b55
	gitlab.cern.ch/lb-experts/golbd v0.2.9
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)

replace gitlab.cern.ch/lb-experts/golbd => ./
