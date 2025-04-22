#!/bin/bash

echo "================================================================================================================"
echo "| NOTE: This script automates the deployment of the Lustre 2.15.4 file system using ZFS as the storage backend "
echo " using offline packages |"
echo "================================================================================================================"
echo ""

set -e
set -o pipefail

PACKAGE_DIR="${package_dir:-./lustre-offline-rpms}"

[ -d "${PACKAGE_DIR}" ] || {
    echo "${PACKAGE_DIR} is not exist!"
    exit 1
}

echo "install all packages under ${PACKAGE_DIR}"
rpm -Uvh "${PACKAGE_DIR}"/*.rpm --force --nodeps || true

echo "install devtools  packages under ${PACKAGE_DIR}/devtools_rpms"
rpm -Uvh "${BASE_DIR}/devtools_rpms/"*.rpm --force --nodeps || true

echo "install kernel-devel and kernel-headers packages under ${PACKAGE_DIR}/kernel-513"
rpm -Uvh "${PACKAGE_DIR}/kernel-513/"*.rpm --force --nodeps || true

echo "install zfs packages under ${PACKAGE_DIR}/zfs"
rpm -Uvh "${PACKAGE_DIR}/zfs/"*.rpm --force --nodeps || true

echo "install lustre packages under ${PACKAGE_DIR}/lustre"
rpm -Uvh "${PACKAGE_DIR}/lustre/"*.rpm --force --nodeps || true

echo "install HA packages under ${PACKAGE_DIR}/ha-rpms"
rpm -Uvh "${BASE_DIR}/ha-rpms/"*.rpm --force --nodeps || true

echo "loading and checking zfs and lustre kernel modules..."
modprobe zfs && echo "zfs modules loaded successfully" || echo "failed to loaded zfs modules"
modprobe lustre && echo "lustre modules loaded successfully" || echo "failed to loaded lustre modules"

