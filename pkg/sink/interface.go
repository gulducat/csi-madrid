package sink

import "github.com/container-storage-interface/spec/lib/go/csi"

/* TODO:
 * context?
 * accept the csi requests to store them, too? they're visible in plugin output
 */

type Sink interface {
	StoreVolume(v *csi.Volume) error
	GetVolume(id string) (*csi.Volume, error)
	ListVolumes() ([]*csi.Volume, error)
	DeleteVolume(id string) error

	StoreSnapshot(s *csi.Snapshot) error
	GetSnapshot(id string) (*csi.Snapshot, error)
	ListSnapshots(volID string) ([]*csi.Snapshot, error)
	DeleteSnapshot(id string) error
}
