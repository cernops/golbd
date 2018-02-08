SNMP lib: SNMP client and trap receiver for golang
--------------------------------
Currently supported operations:
* SNMP v1/v2c/v3 trap receiver with V3 EngineID auto discovery
* SNMP v1/v2c Get, GetMultiple, GetNext, GetBulk, Walk
* SNMP V3     Get, Walk, GetNext

SNMP trap receiver server
--------------------------------
This package includes a helper for running a SNMP trap receiver server. See trapserver.go for more details.
Note that the server does not perform any Community verification. This can be done manually in the OnTrap
function using the provided Trap object.

Using the code
---------------------------------
* The *_test.go files provide good examples of how to use these functions
* Files under examples/ contain the several examples, including an example trap server.

Not supported yet:
------------------
* SNMP Informs receiver
* SNMP v3 GetMultiple, GetBulk (these can be easily implemented since SNMP v3 Walk/Get/GetNext is working)




