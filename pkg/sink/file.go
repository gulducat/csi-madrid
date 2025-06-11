package sink

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/hashicorp/go-hclog"
)

/* TODO:
 * lock?
 * is this thing even useful?
 */

func NewFileSink(log hclog.Logger, dir string) (*FileSink, error) {
	if err := os.MkdirAll(filepath.Join(dir, "volumes"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "snapshots"), 0755); err != nil {
		return nil, err
	}
	return &FileSink{
		dir: dir,
		log: log,
	}, nil
}

var _ Sink = &FileSink{}

type FileSink struct {
	dir string

	log hclog.Logger
}

/* Volumes */

func (f *FileSink) StoreVolume(vol *csi.Volume) error {
	path := filepath.Join(f.dir, "volumes", vol.GetVolumeId())
	f.log.Debug("SaveVolume", "path", path)
	data, err := json.Marshal(vol)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (f *FileSink) GetVolume(id string) (*csi.Volume, error) {
	path := filepath.Join(f.dir, "volumes", id)
	f.log.Debug("GetVolume", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	vol := &csi.Volume{}
	err = json.Unmarshal(data, &vol)
	return vol, err
}

func (f *FileSink) ListVolumes() ([]*csi.Volume, error) {
	vols := []*csi.Volume{}
	dir := filepath.Join(f.dir, "volumes")
	f.log.Debug("ListVolumes", "dir", dir)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		vol, err := f.GetVolume(info.Name())
		if err != nil {
			return err
		}
		vols = append(vols, vol)
		return nil
	})
	return vols, err
}

func (f *FileSink) DeleteVolume(volID string) error {
	path := filepath.Join(f.dir, "volumes", volID)
	f.log.Info("DeleteVolume", "path", path)
	return os.Remove(path)
}

/* Snapshots */

func (f *FileSink) StoreSnapshot(snap *csi.Snapshot) error {
	path := filepath.Join(f.dir, "snapshots", snap.GetSnapshotId())
	f.log.Debug("SaveSnapshot", "path", path)
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (f *FileSink) GetSnapshot(snapID string) (*csi.Snapshot, error) {
	path := filepath.Join(f.dir, "snapshots", snapID)
	f.log.Debug("GetSnapshot", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	snap := &csi.Snapshot{}
	err = json.Unmarshal(data, snap)
	return snap, err
}

func (f *FileSink) DeleteSnapshot(snapID string) error {
	path := filepath.Join(f.dir, "snapshots", snapID)
	f.log.Info("DeleteSnapshot", "path", path)
	return os.Remove(path)
}

func (f *FileSink) ListSnapshots(volID string) ([]*csi.Snapshot, error) {
	snaps := []*csi.Snapshot{}
	dir := filepath.Join(f.dir, "snapshots")
	f.log.Debug("ListSnapshots", "dir", dir)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		snap, err := f.GetSnapshot(info.Name())
		if err != nil {
			return err
		}
		if volID != "" && snap.GetSourceVolumeId() != volID {
			return nil
		}
		snaps = append(snaps, snap)
		return nil
	})
	return snaps, err
}
