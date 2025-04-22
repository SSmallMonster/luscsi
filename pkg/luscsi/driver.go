package luscsi

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	mount "k8s.io/mount-utils"
	"runtime"
)

var (
	controllerServiceCapabilities = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
	}

	volumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}

	nodeServiceCapabilities = []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
	}
)

var (
	DefaultDriverName = "luscsi.luskits.io"
)

const (
	StorageParamMgsAddress = "mgsAddress"
	StorageParamFsName     = "fsName"
	StorageParamSubdir     = "subdir"
)

// DriverOptions defines driver parameters specified in driver deployment
type DriverOptions struct {
	NodeID                     string
	Endpoint                   string
	DriverName                 string
	MountPermissions           uint64
	EnableAzureLustreMockMount bool
	WorkingMountDir            string
	DriverVersion              string
}

// Driver implements all interfaces of CSI drivers
type Driver struct {
	DriverOptions
	cscap   []*csi.ControllerServiceCapability
	nscap   []*csi.NodeServiceCapability
	mounter mount.Interface
}

func (n *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		csc = append(csc, NewControllerServiceCapability(c))
	}
	n.cscap = csc
}

func (n *Driver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	n.nscap = nsc
}

func NewDriver(options *DriverOptions) *Driver {
	d := Driver{
		DriverOptions: *options,
	}

	mounter := mount.New("")
	if runtime.GOOS == "linux" {
		// MounterForceUnmounter is only implemented on Linux now
		mounter = mounter.(mount.MounterForceUnmounter)
	}

	d.mounter = mounter
	return &d
}

func (d *Driver) Run(testMode bool) {
	s := NewNonBlockingGRPCServer()
	s.Start(d.Endpoint, NewIdentifyServer(*d), NewControllerServer(*d), NewNodeServer(*d), testMode)
	s.Wait()
}
