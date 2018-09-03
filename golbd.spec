%global provider	gitlab
%global provider_tld	cern.ch
%global project		lb-experts
%global provider_full %{provider}.%{provider_tld}/%{project}
%global repo		golbd
# %global commit		8c0c623bca8e33f4a9c1289ca965c19d9c6db2b1
%global lbd             lbd

%global import_path	%{provider_full}/%{repo}
%global gopath		%{_datadir}/gocode
# %global shortcommit	%(c=%{commit}; echo ${c:0:7})
%global debug_package	%{nil}

Name:		%{repo}
Version:	0.1
Release:	15
#psaiz: Removing the dist from the release %{?dist}
Summary:	CERN DNS Load Balancer Daemon
License:	ASL 2.0
URL:		https://%{import_path}
# Source:		https://%{import_path}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz
Source:		%{name}-%{version}.tgz
BuildRequires:	systemd
BuildRequires:	golang >= 1.5
ExclusiveArch:	x86_64 

%description
%{summary}

This is a concurrent implementation of the CERN DNS LBD.

The load balancing daemon dynamically handles the list of machines behind a given DNS alias to allow scaling and improve availability.

The Domain Name System (DNS), the defacto standard for name resolution and esential for the network, is an open standard based protocol which allows the use of names instead of IP addresses on the network.
Load balancing is an advanced function that can be provided by DNS, to load balance requests across several machines running the same service by using the same DNS name.

The load balancing server requests each machine for its load status.
The SNMP daemon, gets the request and calls the locally installed metric program, which delivers the load value in SNMP syntax to STDOUT. The SNMP daemon then passes this back to the load balancing server.
The lowest loaded machine names are updated on the DNS servers via the DynDNS mechanism.


%prep
%setup -n %{name}-%{version} -q

%build
mkdir -p src/%{provider_full}
ln -s ../../../ src/%{provider_full}/%{repo}
ln -s src/gitlab.cern.ch . 
(cd src/; ln -s ../vendor/github.com  .)
echo "What do we have"
ls -al src/github.com/reguero/go-snmplib
ls -lR vendor/github.com
echo "AND UNDER SRC"
ls -lR src/github.com

which go
ls -lR src/github.com/
GOPATH=$(pwd):%{gopath} go build %{import_path}

%install
# main package binary
install -d -p %{buildroot}%{_bindir}
install -p -m0755 golbd %{buildroot}%{_bindir}/%{lbd}

# install systemd/sysconfig/logrotate
install -d -m0755 %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m0660 %{lbd}.sysconfig %{buildroot}%{_sysconfdir}/sysconfig/%{lbd} 
install -d -m0755 %{buildroot}%{_unitdir}
install -p -m0644 %{lbd}.service %{buildroot}%{_unitdir}/%{lbd}.service
install -d -m0755 %{buildroot}%{_sysconfdir}/logrotate.d
install -p -m0640 %{lbd}.logrotate %{buildroot}%{_sysconfdir}/logrotate.d/%{lbd}

# create some dirs for logs if needed
install -d -m0755  %{buildroot}/var/log/lb
install -d -m0755  %{buildroot}/var/log/lb/cluster
install -d -m0755  %{buildroot}/var/log/lb/old
install -d -m0755  %{buildroot}/var/log/lb/old/cluster

%check
GOPATH=$(pwd)/:%{gopath} go test %{provider_full}/%{repo}

%post
%systemd_post %{lbd}.service
if [ $1 -eq 1 ] ; then 
        # Initial installation 
        systemctl start lbd.service >/dev/null 2>&1 || : 
fi
if [ $1 -eq 2 ] ; then 
        # Initial installation 
        systemctl try-restart lbd.service >/dev/null 2>&1 || : 
fi


%preun
%systemd_preun %{lbd}.service

%postun
%systemd_postun

%files
%doc LICENSE COPYING README.md 
%attr(755,root,root) %{_bindir}/%{lbd}
%attr(644,root,root) %{_unitdir}/%{lbd}.service
%attr(644,root,root) %config(noreplace) %{_sysconfdir}/sysconfig/%{lbd}
%attr(640,root,root) %{_sysconfdir}/logrotate.d/%{lbd}
%attr(755,root,root) /var/log/lb
%attr(755,root,root) /var/log/lb/cluster
%attr(755,root,root) /var/log/lb/old
%attr(755,root,root) /var/log/lb/old/cluster


%changelog
* Mon Sep  3 2018 Pablo Saiz <pablo.saiz@cern.ch>           - 0.1.15
- Log file in milliseconds
* Mon Jun 18 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1.14
- Using the ip name to check  the host
* Wed Jun 06 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1.13
- Making a single call per host (insead of one call per host per alias)
* Wed Apr 11 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1.11
- Using NoPriv by default
- Detecting more errors in the snmp module
* Wed Mar 14 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1.10
- Changing the snmp module
- Randomizing the list of hosts before sorting them
* Wed Jan 24 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1.9
- Reducing the number of calls to the heartbeat
- Changing the logging

* Tue Nov 28 2017 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.1-7
- Changing category level of errors
* Fri Nov 24 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-6
- heartbeat file needs to be world readable for apache to serve it
* Thu Nov 23 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-5
- add flag to send log to stdout. Fix log name
* Thu Nov 23 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-4
- use lbd drop-in compatible binary and log names
* Mon May 08 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-3
- fix changelog not in descending chronological order
* Mon May 08 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-2
- fix permissions of the systemd service config file
* Mon May 08 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1-1
- point to /usr/bin for the golbd binary in service config
* Sun May 07 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.1.0
- First package for CC7
