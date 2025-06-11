package sink

import (
	"fmt"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func NewMemSink() *MemSink {
	return &MemSink{
		vols:  make(map[string]*csi.Volume),
		snaps: make(map[string]*csi.Snapshot),
	}
}

var _ Sink = &MemSink{}

type MemSink struct {
	lock  sync.RWMutex
	vols  map[string]*csi.Volume
	snaps map[string]*csi.Snapshot
}

func (m *MemSink) StoreVolume(v *csi.Volume) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.vols[v.GetVolumeId()] = v
	return nil
}

func (m *MemSink) GetVolume(id string) (*csi.Volume, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.vols[id]
	if !ok {
		return nil, fmt.Errorf("volume %q not found", id)
	}
	return v, nil
}

func (m *MemSink) ListVolumes() ([]*csi.Volume, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var vols []*csi.Volume
	for _, v := range m.vols {
		vols = append(vols, v)
	}
	return vols, nil
}

func (m *MemSink) DeleteVolume(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.vols, id)
	return nil
}

/* Snapshots */

func (m *MemSink) StoreSnapshot(s *csi.Snapshot) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.snaps[s.GetSnapshotId()] = s
	return nil
}

func (m *MemSink) GetSnapshot(id string) (*csi.Snapshot, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	s, ok := m.snaps[id]
	if !ok {
		return nil, fmt.Errorf("snapshot %q not found", id)
	}
	return s, nil
}

func (m *MemSink) ListSnapshots(volID string) ([]*csi.Snapshot, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var snaps []*csi.Snapshot
	for _, s := range m.snaps {
		if volID == "" || s.GetSourceVolumeId() == volID {
			snaps = append(snaps, s)
		}
	}
	return snaps, nil
}

func (m *MemSink) DeleteSnapshot(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.snaps, id)
	return nil
}
