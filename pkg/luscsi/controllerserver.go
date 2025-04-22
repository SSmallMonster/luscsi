package luscsi

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ControllerServer struct {
	Driver
	csi.ControllerServer
}

func NewControllerServer(driver Driver) *ControllerServer {
	return &ControllerServer{
		Driver: driver,
	}
}

func (d *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	klog.V(2).Infof("CreateVolume called, volumeName: %s", req.GetName())

	if err := checkVolumeRequest(req); err != nil {
		klog.V(2).ErrorS(err, "failed to check request")
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := checkParameters(req.GetParameters()); err != nil {
		klog.V(2).ErrorS(err, "failed to check parameters")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// volume capacity is only supported when lustre is 2.16.0 or higher
	// todo: check lctl version

	volName := req.GetName()
	fsName := req.GetParameters()[StorageParamFsName]
	subdir := req.GetParameters()[StorageParamSubdir]
	mgsAddress := req.GetParameters()[StorageParamMgsAddress]

	if err := d.internalMount(ctx, volName, mgsAddress, fsName, subdir); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount lustre server: %v", err)
	}
	defer func() {
		if err := d.internalUnmount(ctx, volName); err != nil {
			klog.Warningf("failed to unmount lustre server: %v", err)
		}
	}()

	// lustre is mounted now, let's create the volume
	internalPath := path.Join(getInternalMountPath(d.WorkingMountDir, volName), volName)
	if err := os.MkdirAll(internalPath, os.FileMode(d.MountPermissions)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create directory %s: %v", internalPath, err)
	}

	// set directory permissions because of umask problems
	if d.MountPermissions > 0 {
		// Reset directory permissions because of umask problems
		if err := os.Chmod(internalPath, os.FileMode(d.MountPermissions)); err != nil {
			klog.Warningf("failed to chmod subdirectory: %v", err)
		}
	}

	parameters := req.GetParameters()
	setKeyValueInMap(parameters, StorageParamMgsAddress, mgsAddress)
	setKeyValueInMap(parameters, StorageParamFsName, fsName)
	setKeyValueInMap(parameters, StorageParamSubdir, subdir)

	// todo: setup quota
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volName,
			CapacityBytes: 0, // by setting it to zero, Provisioner will use PVC requested size as PV size
			VolumeContext: parameters,
			ContentSource: req.GetVolumeContentSource(),
		},
	}, nil
}

func checkParameters(parameters map[string]string) error {
	if parameters == nil {
		return fmt.Errorf("parameters is empty")
	}

	if parameters[StorageParamMgsAddress] == "" {
		return fmt.Errorf("mgsAddress must be provided")
	}

	if parameters[StorageParamFsName] == "" {
		return fmt.Errorf("fsName must be provided")
	}

	return nil
}

func setKeyValueInMap(m map[string]string, key, value string) {
	if m == nil {
		return
	}
	for k := range m {
		if strings.EqualFold(k, key) {
			m[k] = value
			return
		}
	}
	m[key] = value
}

func (d *ControllerServer) internalUnmount(ctx context.Context, volName string) error {
	targetPath := getInternalMountPath(d.WorkingMountDir, volName)

	// Unmount nfs server at base-dir
	klog.V(2).Infof("internally unmounting %v", targetPath)
	var err error
	forceUnmounter, ok := d.mounter.(mount.MounterForceUnmounter)
	if ok {
		klog.V(2).Infof("force unmount %s on %s", volName, targetPath)
		err = mount.CleanupMountWithForce(targetPath, forceUnmounter, true, 30*time.Second)
	} else {
		err = mount.CleanupMountPoint(targetPath, d.mounter, true)
	}
	if err != nil {
		return status.Errorf(codes.Internal, "failed to unmount target %q: %v", targetPath, err)
	}

	klog.V(2).Infof("internalUnmount: unmount volume %s on %s successfully", volName, targetPath)
	return err
}

func (d *ControllerServer) internalMount(ctx context.Context, volName, mgsAddress, fsName, subdir string) error {
	if volName == "" {
		return status.Error(codes.InvalidArgument,
			"volumeName must be provided")
	}

	if mgsAddress == "" || fsName == "" {
		return status.Error(codes.InvalidArgument,
			"mgsAddress and fsName must be provided")
	}

	sharePath := filepath.Join(mgsAddress, fsName, subdir)
	targetPath := getInternalMountPath(d.WorkingMountDir, volName)
	notMnt, err := d.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, os.FileMode(d.MountPermissions)); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return status.Error(codes.Internal, err.Error())
		}
	}
	if !notMnt {
		klog.V(2).Infof("volumeName %s is already mounted at %s", volName, targetPath)
		return nil
	}

	klog.V(2).Infof("internally mounting %s at %s", sharePath, targetPath)
	execFunc := func() error {
		return d.mounter.Mount(sharePath, targetPath, "lustre", nil)
	}
	timeoutFunc := func() error { return fmt.Errorf("time out") }
	if err := WaitUntilTimeout(90*time.Second, execFunc, timeoutFunc); err != nil {
		if os.IsPermission(err) {
			return status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return status.Error(codes.Internal, err.Error())
	}
	return err
}

func getInternalMountPath(baseDir, volName string) string {
	if baseDir == "" {
		return path.Join("/mnt", volName)
	}
	return path.Join(baseDir, volName)
}

func isVersionGreaterOrEqual(version1, version2 string) bool {
	v1 := strings.Split(version1, ".")
	v2 := strings.Split(version2, ".")

	for i := 0; i < len(v1) && i < len(v2); i++ {
		if v1[i] != v2[i] {
			return v1[i] > v2[i]
		}
	}

	return len(v1) >= len(v2)
}

func checkVolumeRequest(req *csi.CreateVolumeRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "CreateVolumeRequest is nil")
	}

	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "CreateVolume name must be provided")
	}

	if len(req.VolumeCapabilities) == 0 {
		return status.Error(codes.InvalidArgument, "CreateVolume capabilities must be provided")
	}

	if err := validateVolumeCapabilities(req.VolumeCapabilities); err != nil {
		return err
	}

	return nil

}

func validateVolumeCapabilities(capabilities []*csi.VolumeCapability) error {
	for _, capability := range capabilities {
		if capability.GetMount() == nil {
			// Lustre just supports mount type of filesystem
			return status.Error(codes.InvalidArgument,
				"Doesn't support block volume.")
		}
		support := false
		for _, supportedCapability := range volumeCapabilities {
			if capability.GetAccessMode().GetMode() == supportedCapability {
				support = true
				break
			}
		}
		if !support {
			return status.Error(codes.InvalidArgument,
				"Volume doesn't support "+
					capability.GetAccessMode().GetMode().String())
		}
	}
	return nil
}

func (d *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(2).Infof("DeleteVolume called, volumeName: %s", req.GetVolumeId())

	// internal mount lustre to local

	// delete volume
	return nil, nil
}

func (d *ControllerServer) ValidateVolumeCapabilities(
	_ context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, nil
}

func (d *ControllerServer) ControllerGetCapabilities(
	_ context.Context,
	_ *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {
	return nil, nil
}
