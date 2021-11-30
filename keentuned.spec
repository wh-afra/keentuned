%define debug_package %{nil}
%define anolis_release 3

#
# spec file for package golang-keentuned
#

Name:           keentuned
Version:        1.0.0
Release:        %{?anolis_release}%{?dist}
Url:            https://codeup.openanolis.cn/codeup/keentune/keentuned
Summary:        KeenTune tuning tools
License:        MulanPSLv2
Source:         %{name}-%{version}.tar.gz

BuildRoot:      %{_tmppath}/%{name}-%{version}-build
BuildRequires:  go >= 1.13

Vendor:         Alibaba

%description
KeenTune tuning tools rpm package

%prep
%setup -n %{name}-%{version}

%build
go env -w CGO_ENABLED=0
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
cd daemon
go build -ldflags=-linkmode=external -o keentuned
mv -f keentuned ../
cd ../cli
go build -ldflags=-linkmode=external -o keentune
mv -f keentune ../
cd ../

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p ${RPM_BUILD_ROOT}/usr/bin/
mkdir -p ${RPM_BUILD_ROOT}/etc/keentune/
mkdir -p ${RPM_BUILD_ROOT}/etc/keentune/conf/
mkdir -p ${RPM_BUILD_ROOT}/usr/lib/systemd/system/

cp -f keentune ${RPM_BUILD_ROOT}/usr/bin/keentune
cp -f keentuned ${RPM_BUILD_ROOT}/usr/bin/keentuned
cp -rf daemon/examples/. ${RPM_BUILD_ROOT}/etc/keentune
cp -f ./keentuned.conf ${RPM_BUILD_ROOT}/etc/keentune/conf/
cp -f ./keentuned.service ${RPM_BUILD_ROOT}/usr/lib/systemd/system/

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf "$RPM_BUILD_ROOT"
rm -rf $RPM_BUILD_DIR/%{name}-%{version}

%files
%defattr(0444,root,root, 0555)
%attr(0555, root, root) /usr/bin/keentune
%attr(0555, root, root) /usr/bin/keentuned
%{_bindir}/%{name}
%{_bindir}/keentune
%{_sysconfdir}/keentune
%license LICENSE
/usr/lib/systemd/system/keentuned.service

%changelog
* Wed Nov 24 2021 runzhe.wrz <15501019889@126.com> - 1.0.0-3
- add nginx_conf parameter config file

* Wed Nov 10 2021 wenchao <yuxiongkong159@gmail.com> - 1.0.0-2
- use '%license' macro
- update license to MulanPSLv2

* Sun Aug  1 2021 wenchao <yuxiongkong159@gmail.com> - 1.0.0-1
- Init Keentuned.