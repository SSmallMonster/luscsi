#!/bin/bash
# install-os-deps.sh is a helper to utility to install required packages for lustre
#

#set -e

DEBUG=1
function error
{
  echo -e "ERROR: \t$1"
  exit 1
}

function info
{
  echo -e "INFO: \t$1"
}

function debug
{
  if [ $DEBUG -ne 0 ]
  then
    echo -e "DEBUG: \t$1"
  fi
}

function check_error
{    
    if [ ! $? -eq 0 ]
    then
        error "$@"
    fi
}

function online_install_deps
{

    yum install make openssl-devel gcc libtirpc -y 
    debug "install os packages online...."
    info "installing oracle linux 9 repo..."
    tee /etc/yum.repos.d/oracle-baseos.repo <<EOF > /dev/null 2>&1
[ol9_app_latest]
name=Oracle Linux 9 AppStream Latest (x86_64)
baseurl=https://yum.oracle.com/repo/OracleLinux/OL9/appstream/x86_64/
enabled=0
gpgcheck=0
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-oracle
EOF
    #yum makecache > /dev/null 2>&1
    
    yum --disablerepo=* --enablerepo=ol9_app_latest install kernel-devel-5.14.0-427.13.1.el9_4.x86_64 -y > /dev/null 2>&1
    check_error "failed to install kernel-devel-5.14.0-427.13.1.el9_4.x86_64"
    
    yum --disablerepo=* --enablerepo=ol9_app_latest install kernel-headers-5.14.0-427.13.1.el9_4.x86_64 -y > /dev/null 2>&1
    check_error "failed to install kernel-headers-5.14.0-427.13.1.el9_4.x86_64"

    info "installing epel-release..."
    dnf install epel-release -y  > /dev/null 2>&1
    check_error "failed to install epel-release"

    info "installing required packages dkms..."
    yum install dkms -y > /dev/null 2>&1
    check_error "failed to install dkms"

    info "installing required packages python36 libnl3-devel expect python2 bison flex libblkid-devel..."
    yum install libnl3-devel expect bison flex libblkid-devel sysstat -y  > /dev/null 2>&1
    check_error "failed to install python36 libnl3-devel expect python2 bison flex libblkid-devel"

    info "installing zfs-dkms-2.1.15-3.el9 packages..."
    tee /etc/yum.repos.d/zfs-linux.repo <<EOF > /dev/null 2>&1
[zfs_linux]
name=ZFS Linux (x86_64)
baseurl=http://download.zfsonlinux.org/epel/9.4/x86_64/
enabled=0
gpgcheck=0
EOF
    # note: zfs package is only needed when backend storage is zfs rather than ldiskfs
    yum --disablerepo=* --enablerepo=zfs_linux install 	zfs-dkms-2.1.15-3.el9 libnvpair3 libuutil3 libzfs5 libzpool5 zfs-2.1.15 -y  > /dev/null 2>&1
    check_error "failed to install zfs-dkms-2.1.15-3.el9"

    info "installing required packages libyaml-devel..."
    yum --enablerepo=devel install libyaml-devel libmount-devel  -y  > /dev/null 2>&1
    check_error "failed to install libyaml-devel libmount-devel"
    
    info "installing required packages corosync pacemaker pcs..."
    yum --enablerepo=highavailability install -y corosync pacemaker pcs > /dev/null 2>&1
    check_error "failed to install corosync pacemaker pcs"
}

function main
{
    echo "================================================================================================================"
    echo "| NOTE: This script automates the deployment of the Lustre 2.16.1 file system using ZFS as the storage backend. |"
    echo "================================================================================================================"
    echo ""

    ! grep -q "Rocky Linux release 9.4" /etc/redhat-release && error "os not match, this script is only for Rocky Linux release 8.9!
        download link: https://dl.rockylinux.org/vault/rocky/8.9/isos/x86_64/Rocky-8.9-x86_64-minimal.iso
    "
    
    [ "5.14.0-427.13.1.el9_4.x86_64" != $(uname -r) ] && error "kernel version(`uname -r`) not match, required kernel is 5.14.0-427.13.1.el9_4.x86_64!
        download link: https://dl.rockylinux.org/vault/rocky/8.9/isos/x86_64/Rocky-8.9-x86_64-minimal.iso
    "

    info "os info: `cat /etc/redhat-release` `uname -r` check passed..."

    online_install_deps
}

main
