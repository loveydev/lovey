package lovey

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"time"
)

var (
	ErrRootMetadataNotFound     = errors.New("root metadata file not found")
	ErrRollbackAttack           = errors.New("rollback attack detected")
	ErrExpiredRootMetadata      = errors.New("latest root metadata file has expired")
	ErrExpiredTimestampMetadata = errors.New("latest timestamp metadata file has expired")
	ErrExpiredSnapshotMetadata  = errors.New("latest snapshot metadata file has expired")
)

type Client struct {
	c *http.Client

	files          fs.FS
	trustedFileDir string

	root      Root
	timestamp *Timestamp
	snapshot  *Snapshot

	fixedStartTime time.Time

	maxRootMetadataSize       int64
	maxTimestampMetadataSize  int64
	maxNumRootMetadataFetches int
}

func (c *Client) loadRoot(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *Client) persistRoot(ctx context.Context, root Root) error {
	return errors.New("not implemented")
}

func (c *Client) loadTimestamp(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *Client) persistTimestamp(ctx context.Context, ts Timestamp) error {
	return errors.New("not implemented")
}

func (c *Client) deleteTimestampFile(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *Client) loadSnapshot(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *Client) persistSnapshot(ctx context.Context, s Snapshot) error {
	return errors.New("not implemented")
}

func (c *Client) deleteSnapshotFile(ctx context.Context) error {
	return errors.New("not implemented")
}

func (c *Client) buildRootURL(version int64) (string, error) {
	return "", errors.New("not implemented")
}

func (c *Client) buildTimestampURL() (string, error) {
	return "", errors.New("not implemented")
}

func (c *Client) buildSnapshotURL() (string, error) {
	return "", errors.New("not implemented")
}

func (c *Client) fetchRoot(ctx context.Context, u string) (Root, error) {
	return Root{}, errors.New("not implemented")
}

func (c *Client) fetchTimestamp(ctx context.Context, u string) (Timestamp, error) {
	return Timestamp{}, errors.New("not implemented")
}

func (c *Client) fetchSnapshot(ctx context.Context, u string) (Snapshot, error) {
	return Snapshot{}, errors.New("not implemented")
}

func (c *Client) updateRoot(ctx context.Context) error {
	c.fixedStartTime = time.Now()
	err := c.loadRoot(ctx)
	if err != nil {
		return err
	}
	origRoles := c.root.Roles
	for i := 0; i < c.maxNumRootMetadataFetches; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		u, err := c.buildRootURL(c.root.Version + 1)
		if err != nil {
			return err
		}
		newRoot, err := c.fetchRoot(ctx, u)
		if err == ErrRootMetadataNotFound {
			break
		}
		if err != nil {
			return err
		}
		err = newRoot.checkSignatures(ctx, c.root.Keys, c.root.Roles.Root)
		if err != nil {
			return err
		}
		err = newRoot.checkSignatures(ctx, newRoot.Keys, newRoot.Roles.Root)
		if err != nil {
			return err
		}
		if newRoot.Version < c.root.Version {
			return ErrRollbackAttack
		}
		err = c.persistRoot(ctx, newRoot)
		if err != nil {
			return err
		}
		c.root = newRoot
	}

	if !c.root.Expires.After(c.fixedStartTime) {
		return ErrExpiredRootMetadata
	}

	if origRoles.Snapshot.rotated(c.root.Roles.Snapshot) ||
		origRoles.Timestamp.rotated(c.root.Roles.Timestamp) {
		err = c.deleteTimestampFile(ctx)
		if err != nil {
			return err
		}
		err = c.deleteSnapshotFile(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) updateTimestamp(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := c.loadTimestamp(ctx)
	if err != nil {
		return err
	}

	u, err := c.buildTimestampURL()
	if err != nil {
		return err
	}
	newTimestamp, err := c.fetchTimestamp(ctx, u)
	if err != nil {
		return err
	}
	err = newTimestamp.checkSignatures(ctx, c.root.Keys, c.root.Roles.Timestamp)
	if err != nil {
		return err
	}
	if c.timestamp != nil {
		if newTimestamp.Version < c.timestamp.Version {
			return ErrRollbackAttack
		}
		trustedSnapshot, ok := c.timestamp.Meta["snapshot.json"]
		if !ok {
			return errors.New("no snapshot.json recorded in trusted timestamp meta")
		}
		newSnapshot := newTimestamp.Meta["snapshot.json"]
		if !ok {
			return errors.New("no snapshot.json recorded in candidate timestamp meta")
		}
		if trustedSnapshot.Version > newSnapshot.Version {
			return errors.New("candidate timestamp's meta has a snapshot.json with a lower version")
		}
	}
	if !newTimestamp.Expires.After(c.fixedStartTime) {
		return ErrExpiredTimestampMetadata
	}
	err = c.persistTimestamp(ctx, newTimestamp)
	if err != nil {
		return err
	}
	c.timestamp = &newTimestamp
	return nil
}

func (c *Client) updateSnapshot(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := c.loadSnapshot(ctx)
	if err != nil {
		return err
	}

	u, err := c.buildSnapshotURL()
	if err != nil {
		return err
	}
	newSnapshot, err := c.fetchSnapshot(ctx, u)
	if err != nil {
		return err
	}
	if c.timestamp == nil {
		return errors.New("no trusted timestamp metadata")
	}
	tsSnapshot, ok := c.timestamp.Meta["snapshot.json"]
	if !ok {
		return errors.New("no snapshot.json recorded in trusted timestamp meta")
	}
	err = newSnapshot.checkHashes(tsSnapshot.Hashes)
	if err != nil {
		return err
	}
	err = newSnapshot.checkSignatures(ctx, c.root.Keys, c.root.Roles.Snapshot)
	if err != nil {
		return err
	}
	if newSnapshot.Version != tsSnapshot.Version {
		return errors.New("candidate snapshot file's version didn't match version in the timestamp file")
	}
	if c.snapshot != nil {
		err := newSnapshot.checkVersions(ctx, c.snapshot.Meta)
		if err != nil {
			return err
		}
		err = newSnapshot.checkTargets(ctx, c.snapshot.Meta)
		if err != nil {
			return err
		}
	}
	if newSnapshot.Expires.After(c.fixedStartTime) {
		return ErrExpiredSnapshotMetadata
	}
	err = c.persistSnapshot(ctx, newSnapshot)
	if err != nil {
		return err
	}
	c.snapshot = &newSnapshot
	return nil
}
