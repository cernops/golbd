module lb-experts/golbd

go 1.17

require (
	github.com/miekg/dns v0.0.0-20160605072344-799de7044d95
	github.com/reguero/go-snmplib v0.0.0-20181019092238-e566f5619b55
	gitlab.cern.ch/lb-experts/golbd v0.2.9
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace gitlab.cern.ch/lb-experts/golbd => ./
