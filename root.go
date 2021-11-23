package lovey

import (
	"context"
	"errors"
	"time"
)

type Root struct {
	SpecVersion        string
	ConsistentSnapshot bool
	Version            int64
	Expires            time.Time
	Keys               map[string]Key
	Roles              RootRoles
}

func (r Root) checkSignatures(ctx context.Context, keys map[string]Key, root RootRole) error {
	return errors.New("not implemented")
}

type RootRoles struct {
	Root      RootRole
	Snapshot  RootRole
	Targets   RootRole
	Timestamp RootRole
	Mirrors   *RootRole
}

type RootRole struct {
	KeyIDs    []string
	Threshold int64
}

func (r RootRole) rotated(newest RootRole) bool {
	for _, id := range r.KeyIDs {
		var found bool
		for _, n := range newest.KeyIDs {
			if id == n {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}
