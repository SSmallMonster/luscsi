package luscsi

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeServer struct {
	Driver
	csi.NodeServer
}

func NewNodeServer(driver Driver) *NodeServer {
	return &NodeServer{
		Driver: driver,
	}
}

func (d *NodeServer) NodePublishVolume(
	_ context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {
	return nil, nil
}

func (d *NodeServer) NodeUnpublishVolume(
	_ context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, nil
}

func (d *NodeServer) NodeStageVolume(_ context.Context, _ *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// Staging and Unstaging is not able to be supported with how Lustre is mounted
//
// See NodeStageVolume for more details
func (d *NodeServer) NodeUnstageVolume(_ context.Context, _ *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *NodeServer) NodeGetCapabilities(
	_ context.Context, _ *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: d.nscap,
	}, nil
}

// NodeGetInfo return info of the node on which this plugin is running
func (d *NodeServer) NodeGetInfo(
	_ context.Context,
	_ *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: d.NodeID,
	}, nil
}

func (d *NodeServer) NodeGetVolumeStats(
	_ context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
