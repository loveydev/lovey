package lovey

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Snapshot struct {
	SpecVersion string
	Version     int64
	Expires     time.Time
	Meta        map[string]SnapshotMetaFile
}

func (s Snapshot) checkHashes(hashes map[string]string) error {
	return errors.New("not implemented")
}

func (s Snapshot) checkSignatures(ctx context.Context, keys map[string]Key, role RootRole) error {
	return errors.New("not implemented")
}

func (s Snapshot) checkVersions(ctx context.Context, trusted map[string]SnapshotMetaFile) error {
	return errors.New("not implemented")
}

func (s Snapshot) checkTargets(ctx context.Context, trusted map[string]SnapshotMetaFile) error {
	for name := range trusted {
		if _, ok := s.Meta[name]; !ok {
			return fmt.Errorf("snapshot meta for target %q not found in new snapshot meta file", name)
		}
	}
	return nil
}

type SnapshotMetaFile struct {
	Version int64
	Length  *int64
	Hashes  map[string]string
}
