%define         debug_package %{nil}

Name:		tap-ceph-broker
Version:	%{pkg_version}
Release:	1%{?dist}
Summary:	Application for managing CEPH resources needed by TAP core services

License:	ASL 2.0
URL:		https://github.com/trustedanalytics/tap-ceph-broker
Source0:	tap-ceph-broker

%description
%{summary}

%prep

%build

%install
install -D -p -m 755 %{SOURCE0}/application/tap-ceph-broker %{buildroot}%{_bindir}/tap-ceph-broker
install -D -p -m 644 %{SOURCE0}/tap-ceph-broker.service %{buildroot}%{_unitdir}/tap-ceph-broker.service
install -D -p -m 644 %{SOURCE0}/tap-ceph-broker.conf %{buildroot}%{_sysconfdir}/sysconfig/tap-ceph-broker

%post
%systemd_post tap-ceph-broker.service

%postun
%systemd_postun_with_restart tap-ceph-broker.service

%files
%{_bindir}/tap-ceph-broker
%{_unitdir}/tap-ceph-broker.service
%config(noreplace) %{_sysconfdir}/sysconfig/tap-ceph-broker

%changelog
* Mon Oct 10 2016 - Mariusz Klonowski <mariusz.klonowski@intel.com>
- Initial Build
