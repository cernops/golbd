%global provider	github
%global provider_tld	com
%global project		cernops
%global repo		golbd
# %global commit		8c0c623bca8e33f4a9c1289ca965c19d9c6db2b1

%global import_path	%{provider}.%{provider_tld}/%{project}/%{repo}
%global gopath		%{_datadir}/gocode
# %global shortcommit	%(c=%{commit}; echo ${c:0:7})
%global debug_package	%{nil}

Name:		%{repo}
Version:	0.1
Release:	1%{?dist}
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
mkdir _build

pushd _build
  mkdir -p src/%{provider}.%{provider_tld}/%{project}
  ln -s $(dirs +1 -l) src/%{import_path}
popd

GOPATH=$(pwd)/_build:%{gopath} go build %{import_path}

%install
# main package binary
install -d -p %{buildroot}%{_bindir}
install -p -m0755 golbd %{buildroot}%{_bindir}

# install systemd/sysconfig/logrotate
install -d -m0755 %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m0660 %{name}.sysconfig %{buildroot}%{_sysconfdir}/sysconfig/%{name} 
install -d -m0755 %{buildroot}%{_unitdir}
install -p -m0644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service
install -d -m0755 %{buildroot}%{_sysconfdir}/logrotate.d
install -p -m0640 %{name}.logrotate %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

# create some dirs for logs if needed
install -d -m0755  %{buildroot}/var/log/lb
install -d -m0755  %{buildroot}/var/log/lb/cluster
install -d -m0755  %{buildroot}/var/log/lb/old
install -d -m0755  %{buildroot}/var/log/lb/old/cluster

%check
GOPATH=$(pwd)/_build:%{gopath} go test github.com/cernops/golbd

%post
%systemd_post golbd.service

%preun
%systemd_preun golbd.service

%postun
%systemd_postun

%files
%doc LICENSE COPYING README.md 
%attr(755,root,root) %{_bindir}/golbd
%attr(755,root,root) %{_unitdir}/%{name}.service
%attr(644,root,root) %config(noreplace) %{_sysconfdir}/sysconfig/%{name}
%attr(640,root,root) %{_sysconfdir}/logrotate.d/%{name}
%attr(755,root,root) /var/log/lb
%attr(755,root,root) /var/log/lb/cluster
%attr(755,root,root) /var/log/lb/old
%attr(755,root,root) /var/log/lb/old/cluster


%changelog
* Sun May 07 2017 Ignacio Reguero <Ignacio.Reguero@cern.ch> - 0.0.1-0.0.gitf50fc79
- First package for CC7
