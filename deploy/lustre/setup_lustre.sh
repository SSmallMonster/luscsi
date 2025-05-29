#!/bin/bash

CONFIG_FILE=${1:-example_lustre.cfg}
[ ! -f "$CONFIG_FILE" ] && echo "config $CONFIG_FILE is not exist!" && exit 1
source "$CONFIG_FILE"

[ "$(id -u)" != "0" ] && echo "root is required!" && exit 1

function create_zpool() {
    local pool=$1
    local disks=($2)
    local type=${3:-stripe}

    if ! zpool list "$pool" &> /dev/null; then
        case $type in
            stripe) zpool create "$pool" "${disks[@]}" ;;
            mirror) zpool create "$pool" mirror "${disks[@]}" ;;
            raidz1) zpool create "$pool" raidz1 "${disks[@]}" ;;
            *) echo "unknown zpool type: $type"; exit 1 ;;
        esac
        echo "zpool: $pool is created with type: $type and disks: ${disks[*]}"
    fi
}

function setup_lustre() {
    case $1 in
        mgs)
	          fsName=$(zfs get -H -o value lustre:fsname "$MGS_POOL"/"$MGT_TARGET" 2> /dev/null)
	          { [ "$fsName" == "" ] && mkfs.lustre --mgs --mdt --fsname="$FSNAME" --backfstype=zfs --mgsnode="$MGS_NODE" --servicenode="$MGS_NODE" --index="$MGS_INDEX" "$MGS_POOL"/"$MGT_TARGET"; } || \
	          { echo "$MGS_POOL/$MGT_TARGET is already formated as lustre filesystem"; }

	          mkdir -p "$MGS_MOUNT"
	          [ ! "$(mountpoint "$MGS_MOUNT" >/dev/null && findmnt -no FSTYPE "$MGS_MOUNT")" == "lustre" ] && \
	          mount -t lustre "$MGS_POOL"/"$MGT_TARGET" "$MGS_MOUNT"
            ;;

        ost)
            for ost_conf in "${OST_CONFIGS[@]}"; do
		          IFS=':' read -r pool target lindex disks otype mountdir <<< "$ost_conf"
		          fsName=$(zfs get -H -o value lustre:fsname "$pool"/"$target" 2> /dev/null)
		          { [ "$fsName" == "" ] && mkfs.lustre --ost --fsname="$FSNAME" --backfstype=zfs --servicenode="$MGS_NODE" --mgsnode="$MGS_NODE" --index="$lindex" "$pool"/"$target"; }  || \
		          { echo "$pool/$target is already formated as lustre filesystem"; }

              mkdir -p "$mountdir"
		          [ ! "$(mountpoint "$mountdir" >/dev/null && findmnt -no FSTYPE "$mountdir")" == "lustre" ] && \
              mount -t lustre "$pool"/"$target" "$mountdir"
            done
            ;;
    esac
}

function check_servers(){
	  test_mount="/mnt/lustre~dir" && mkdir -p $test_mount

	  [ "$(mountpoint $test_mount >/dev/null && findmnt -no FSTYPE $test_mount)" == "" ] && \
	  { mount.lustre "$MGS_NODE":/"$FSNAME" $test_mount || { echo "failed to mount.lustre $MGS_NODE:$FSNAME at $test_mount" && exit 1; } }

    echo "lustre is mounted at $test_mount, now checking servers..."
	  lfs check servers || { echo "failed to check lustre servers" && exit 1; }
	  echo "lustre is setting up successfully"
}

function init_service() {
    # lustre service
    modprobe zfs
    modprobe lustre
    modprobe lnet
    
    # os service 
    systemctl disable firewalld
}

function init_network() {
    lnetctl lnet configure
    # check lnid exist or not by ping itself
    { lctl which_nid "$MGS_NODE" >/dev/null && echo "nid $MGS_NODE is already created"; } || \
    # configure lnet using lnetctl
    { IFS="@" read -r _ lnet <<< "$MGS_NODE" && lnetctl net add --net "$lnet" --if "$NETWORK"; }
}

function main() {
    init_service
    init_network

    # prepare zpool for lustre targets
    [ ! -f /etc/hostid ] && zgenhostid
    create_zpool "$MGS_POOL" "$MGS_DISKS" "$MGS_TYPE"
    for ost_conf in "${OST_CONFIGS[@]}"; do
	  IFS=':' read -r pool target _ disks otype mountdir <<< "$ost_conf"
        create_zpool "$pool" "$disks" "$otype"
    done

    # format lustre targets using zfs dataset and bring it up
    setup_lustre mgs
    setup_lustre ost

    # make sure lustre is up and can be mounted
    check_servers
}

main
