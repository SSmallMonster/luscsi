package luscsi

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type IdentifyServer struct {
	Driver
	csi.IdentityServer
}

func NewIdentifyServer(driver Driver) *IdentifyServer {
	return &IdentifyServer{
		Driver: driver,
	}
}

// GetPluginInfo return the version and name of the plugin
func (d *IdentifyServer) GetPluginInfo(_ context.Context, _ *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	if d.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if d.DriverVersion == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          d.DriverName,
		VendorVersion: d.DriverVersion,
	}, nil
}

// Probe check whether the plugin is running or not.
// This method does not need to return anything.
// Currently the spec does not dictate what you should return either.
// Hence, return an empty response
func (d *IdentifyServer) Probe(_ context.Context, _ *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{Ready: &wrapperspb.BoolValue{Value: true}}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin
func (d *IdentifyServer) GetPluginCapabilities(_ context.Context, _ *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
