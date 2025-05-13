package main

import (
	"flag"
	"github.com/luskits/luscsi/pkg/luscsi"
	klog "k8s.io/klog/v2"
	"os"
)

var (
	endpoint         = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID           = flag.String("nodeid", "", "node id")
	mountPermissions = flag.Uint64("mount-permissions", 0, "mounted folder permissions")
	driverName       = flag.String("drivername", luscsi.DefaultDriverName, "name of the driver")
	workingMountDir  = flag.String("working-mount-dir", "/mnt", "working directory for provisioner to mount nfs shares temporarily")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	if *nodeID == "" {
		klog.Fatal("nodeid is empty")
	}

	handle()
	os.Exit(0)
}

func handle() {
	driverOptions := luscsi.DriverOptions{
		NodeID:           *nodeID,
		DriverName:       *driverName,
		Endpoint:         *endpoint,
		MountPermissions: *mountPermissions,
		WorkingMountDir:  *workingMountDir,
	}
	d := luscsi.NewDriver(&driverOptions)
	d.Run(false)
}
