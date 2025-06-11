package server

import (
	"context"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/hashicorp/go-hclog"
)

/* TODO:
 * sink
 * panics
 * capabilities?
 * toplogy?
 */

type NodeServer struct {
	nodeID string

	log hclog.Logger

	csi.NodeServer
}

func (n *NodeServer) NodeExpandVolume(_ context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	n.log.Debug("NodeExpandVolume", "req", req)
	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
	}, nil
}

func (n *NodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			// {
			// 	Type: &csi.NodeServiceCapability_Rpc{
			// 		Rpc: &csi.NodeServiceCapability_RPC{
			// 			Type: csi.NodeServiceCapability_RPC_UNKNOWN,
			// 		},
			// 	},
			// },
		},
	}, nil
}

// required regardless of capabilities
func (n *NodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: n.nodeID,
		// MaxVolumesPerNode:  0,
		// AccessibleTopology: &csi.Topology{},
	}, nil
}

func (n *NodeServer) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("unimplemented")
}

func (n *NodeServer) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	n.log.Debug("NodePublishVolume", "req", req)
	err := os.MkdirAll(req.GetTargetPath(), os.ModeDir)
	return &csi.NodePublishVolumeResponse{}, err
}

func (n *NodeServer) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	panic("unimplemented")
}

func (n *NodeServer) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	n.log.Debug("NodeUnpublishVolume", "req", req)
	err := os.RemoveAll(req.GetTargetPath())
	return &csi.NodeUnpublishVolumeResponse{}, err
}

func (n *NodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("unimplemented")
}
