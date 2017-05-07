# Description of the LBD

## Presentation of the Service
The load balancing service dynamically handles the list of machines behind a given DNS alias to allow scaling and improve availability by allowing several nodes to be presented behind a single name. It is one of the technologies enabling deployment of large scale applications on cloud resources.

The Domain Name System (DNS) is a naming system for computers, or other resources. DNS is essential for Internet. It is an internet-standard protocol that allows to use names instead of IP addresses. Load balancing is an advanced function that can be provided by DNS, to distribute requests across several machines running the same service by using the same DNS name.

The LBD DNS Load Balancer has been developed at CERN as a cost-effective way to handle applications accepting the DNS timing constraints and not requiring affinity (also known as persistence or sticky sessions). 
It is currently (May 2017) used by over 500 services on CERN the site with two small VMs acting as LBD master and slave. The alias member nodes have configured a Simple Network Management Protocol (SNMP) agent that communicates with the lbclient program. The lbclient provides a load metric number used by the LBD server to determine the subset of nodes from the set whose IP is to be presented. Here there is an overview of the service.

The LBD server periodically gets a load metric from the alias member nodes using SNMP and uses the information to update the A (IPV4) and AAAA (IPV6) records for a DNS delegated zone that corresponds to the alias using Dynamic DNS (see RFC2136). The period ("polling_interval") is 5 minutes by default.

The "best_hosts" parameter defines the number of nodes exposed by the load balanced alias at any time, so the LBD server takes the "best_hosts" number of nodes that are the least loaded and updates the DNS A and AAAA records if needed for their IP addresses to be exposed by the alias.  If best_hosts has special value of -1, to expose the IP addresses of all alias members. This is used, for instance, on CERN message brokers where it we have to consume from all alias members.

There is a caveat when more than one IP address is presented, some resolvers will bias deterministically the order of the list of IPs. For instance in SLC5 this is due to the bug of getaddrinfo() in glibc described in the following twiki: https://twiki.cern.ch/twiki/bin/view/LinuxSupport/GlibcDnsLoadBalancing

The LBD slave does like the LBD master, i.e: periodically gets a load metric from the alias member nodes, however, it only updates the DNS delegated zone when it loses contact with the LBD master. This is verified by trying to get a file with a "heartbeat" from a web server on the LBD master.

The lbclient provides a built-in load metric. Alternative load metrics can be configured by combining several Lemon metrics and constants. Health monitoring checks can also be configured for the alias members to be taken out of the alias when certain condition is triggered. A typical example is the check of the Roger state so that the node is taken out when the appstate is not 'production'. As well as several built-in checks you may also configure additional ones using Lemon metrics. You can also use the return code of an arbitrary program (or script) as a check. If the node is in working state the load metric is an integer greater than 0. If the load metric is 0 or lower than 0, it means that the machine is not available.

## Additional Information

For more information, you can check the user documentation located in [configdocs](http://configdocs.web.cern.ch/configdocs/dnslb/index.html)
