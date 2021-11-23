package lovey

import (
	"context"
	"errors"
	"time"
)

type Timestamp struct {
	SpecVersion string
	Version     int64
	Expires     time.Time
	Meta        map[string]TimestampMetaFile
}

func (t Timestamp) checkSignatures(ctx context.Context, keys map[string]Key, ts RootRole) error {
	return errors.New("not implemented")
}

type TimestampMetaFile struct {
	Version int64
	Length  *int64
	Hashes  map[string]string
}
