package server

import (
	"context"
	"fmt"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/gulducat/csi-madrid/pkg/sink"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/* TODO:
 * unimplemented
 * volume response Status
 * topology
 */

type ControllerServer struct {
	sink sink.Sink

	log hclog.Logger

	csi.ControllerServer
}

func (c *ControllerServer) ControllerExpandVolume(_ context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	c.log.Debug("ControllerExpandVolume", "req", req)
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         req.GetCapacityRange().GetRequiredBytes(),
		NodeExpansionRequired: true, // hah
	}, nil
}

func (c *ControllerServer) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	caps := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
	}
	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: make([]*csi.ControllerServiceCapability, len(caps)),
	}
	for i, cap := range caps {
		resp.Capabilities[i] = &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				}}}
	}
	return resp, nil
}

func (c *ControllerServer) ControllerGetVolume(_ context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	vol, err := c.sink.GetVolume(req.GetVolumeId())
	if err != nil {
		return nil, err
	}
	return &csi.ControllerGetVolumeResponse{
		Volume: vol,
		Status: nil, // TODO
	}, nil
}

func (c *ControllerServer) ControllerModifyVolume(context.Context, *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	panic("unimplemented")
}

func (c *ControllerServer) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	panic("unimplemented")
}

func (c *ControllerServer) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	panic("unimplemented")
}

func (c *ControllerServer) CreateSnapshot(_ context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	source := req.GetSourceVolumeId()
	vol, err := c.sink.GetVolume(source)
	if err != nil {
		return nil, fmt.Errorf("error getting volume %q for snapshot: %w", source, err)
	}
	snap := &csi.Snapshot{
		SnapshotId:      req.GetName(),
		SourceVolumeId:  vol.GetVolumeId(),
		CreationTime:    timestamppb.New(time.Now()),
		ReadyToUse:      true,
		SizeBytes:       0,
		GroupSnapshotId: "",
	}
	err = c.sink.StoreSnapshot(snap)
	return &csi.CreateSnapshotResponse{Snapshot: snap}, err
}

func (c *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	vol := &csi.Volume{
		VolumeId:      req.GetName(),
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
		VolumeContext: req.GetParameters(),
		ContentSource: req.GetVolumeContentSource(),
		// AccessibleTopology: []*csi.Topology{},
	}
	err := c.sink.StoreVolume(vol)
	return &csi.CreateVolumeResponse{Volume: vol}, err
}

func (c *ControllerServer) DeleteSnapshot(_ context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	err := c.sink.DeleteSnapshot(req.GetSnapshotId())
	return &csi.DeleteSnapshotResponse{}, err
}

func (c *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	err := c.sink.DeleteVolume(req.GetVolumeId())
	return &csi.DeleteVolumeResponse{}, err
}

func (c *ControllerServer) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	panic("unimplemented")
}

func (c *ControllerServer) ListSnapshots(_ context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	snaps, err := c.sink.ListSnapshots(req.GetSourceVolumeId())
	if err != nil {
		return nil, err
	}
	resp := &csi.ListSnapshotsResponse{
		Entries: make([]*csi.ListSnapshotsResponse_Entry, len(snaps)),
	}
	for i, snap := range snaps {
		resp.Entries[i] = &csi.ListSnapshotsResponse_Entry{Snapshot: snap}
	}
	return resp, nil
}

func (c *ControllerServer) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	vols, err := c.sink.ListVolumes()
	if err != nil {
		return nil, err
	}
	resp := &csi.ListVolumesResponse{
		Entries: make([]*csi.ListVolumesResponse_Entry, len(vols)),
	}
	for i, v := range vols {
		resp.Entries[i] = &csi.ListVolumesResponse_Entry{
			Volume: v,
			Status: nil, // TODO
		}
	}
	return resp, nil
}

func (c *ControllerServer) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	panic("unimplemented")
}
