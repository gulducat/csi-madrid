package sink

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

/* TODO:
 * the locking in here is sketchy
 */

func NewNomadSink(log hclog.Logger, vars *api.Variables, basePath, region, namespace string) *NomadSink {
	log.Info("new nomad sink", "basePath", basePath, "region", region, "namespace", namespace)
	return &NomadSink{
		vars:      vars,
		basePath:  basePath,
		region:    region,
		namespace: namespace,
		log:       log,
	}
}

var _ Sink = &NomadSink{}

type NomadSink struct {
	vars     *api.Variables
	basePath string

	region    string
	namespace string

	log hclog.Logger
}

/* helpers */

func (n *NomadSink) lock() (func(), error) { // TODO
	unlock := func() {}
	path := n.basePath + "/lock"
	n.log.Debug("acquiring lock", "path", path)
	nVar := &api.Variable{
		Namespace: n.namespace,
		Path:      path,
		Items:     nil,
		Lock: &api.VariableLock{
			ID:        "csi-madrid",
			TTL:       "15s",
			LockDelay: "30s",
		},
	}
	if _, _, err := n.vars.AcquireLock(nVar, nil); err != nil {
		return unlock, fmt.Errorf("failed to acquire lock: %w", err)
	}
	unlock = func() {
		n.log.Debug("releasing lock", "path", path)
		if _, _, err := n.vars.ReleaseLock(nVar, nil); err != nil {
			n.log.Warn("failed to release lock", "error", err)
		}
	}
	return unlock, nil
}

func (n *NomadSink) runLocked(fn func() error) error {
	// TODO: retry
	unlock, err := n.lock()
	if err != nil {
		return err
	}
	defer unlock()
	return fn()
}

/* Volumes */

func (n *NomadSink) volPath(id string) string {
	return n.basePath + "/volumes/" + id
}

func (n *NomadSink) StoreVolume(v *csi.Volume) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return n.runLocked(func() error {
		path := n.volPath(v.GetVolumeId())
		n.log.Info("saving volume", "path", path, "json", string(data))

		_, _, err = n.vars.Create(&api.Variable{
			Namespace: n.namespace,
			Path:      path,
			Items:     map[string]string{"json": string(data)},
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to create nomad var: %w", err)
		}
		return nil
	})
}

func (n *NomadSink) GetVolume(id string) (*csi.Volume, error) {
	vol := &csi.Volume{}
	err := n.runLocked(func() error {
		path := n.volPath(id)
		n.log.Info("getting volume", "path", path)

		items, _, err := n.vars.GetVariableItems(path, nil)
		if err != nil {
			return fmt.Errorf("failed to get nomad var: %w", err)
		}
		n.log.Debug("got variable items", "path", path, "items", items)
		volJson := items["json"]
		return json.Unmarshal([]byte(volJson), vol)
	})
	return vol, err
}

func (n *NomadSink) ListVolumes() ([]*csi.Volume, error) {
	return nil, nil // TODO?
}

func (n *NomadSink) DeleteVolume(id string) error {
	return n.runLocked(func() error {
		path := n.volPath(id)
		n.log.Info("deleting volume", "path", path)

		_, err := n.vars.Delete(path, nil)
		if err != nil {
			return fmt.Errorf("failed to delete nomad var: %w", err)
		}
		return nil
	})
}

/* Snapshots */

func (n *NomadSink) snapPath(id string) string {
	return n.basePath + "/snapshots/" + id
}

func (n *NomadSink) StoreSnapshot(s *csi.Snapshot) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return n.runLocked(func() error {
		path := n.snapPath(s.GetSnapshotId())
		n.log.Info("saving snapshot", "path", path, "json", string(data))

		_, _, err = n.vars.Create(&api.Variable{
			Namespace: n.namespace,
			Path:      path,
			Items:     map[string]string{"json": string(data)},
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to create nomad var: %w", err)
		}
		return nil
	})
}

func (n *NomadSink) GetSnapshot(id string) (*csi.Snapshot, error) {
	snap := &csi.Snapshot{}
	err := n.runLocked(func() error {
		path := n.snapPath(id)
		n.log.Info("getting snapshot", "path", path)

		items, _, err := n.vars.GetVariableItems(path, nil)
		if err != nil {
			return fmt.Errorf("failed to get nomad var: %w", err)
		}
		n.log.Debug("got var", "items", items)
		snapJson := items["json"]
		return json.Unmarshal([]byte(snapJson), snap)
	})
	return snap, err
}

func (n *NomadSink) ListSnapshots(volID string) ([]*csi.Snapshot, error) {
	// TODO: lock better - can't lock in GetSnapshot while lock is held for ListSnapshots
	snaps := []*csi.Snapshot{}
	var varMeta []*api.VariableMetadata
	err := n.runLocked(func() error {
		path := n.snapPath("")
		n.log.Info("listing snapshots", "path", path)

		var err error
		varMeta, _, err = n.vars.PrefixList(path, nil)
		return err
	})
	if err != nil {
		return snaps, err
	}
	for _, vm := range varMeta {
		pParts := strings.Split(vm.Path, "/")
		snapID := pParts[len(pParts)-1]
		n.log.Debug("list response", "pparts", pParts)
		snap, err := n.GetSnapshot(snapID)
		if err != nil {
			n.log.Error(err.Error())
			continue
		}
		snaps = append(snaps, snap)
	}
	return snaps, nil
}

func (n *NomadSink) DeleteSnapshot(id string) error {
	return n.runLocked(func() error {
		path := n.snapPath(id)
		n.log.Info("deleting snapshot", "path", path)

		_, err := n.vars.Delete(path, nil)
		if err != nil {
			return fmt.Errorf("failed to delete nomad var: %w", err)
		}
		return nil
	})
}
