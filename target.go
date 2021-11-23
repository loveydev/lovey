package lovey

import "time"

type Target struct {
	SpecVersion string
	Version     int64
	Expires     time.Time
	Targets     map[string]TargetMeta
	Delegations map[string]TargetDelegation
}

type TargetMeta struct {
	Length int64
	Hashes map[string]string
	Custom map[string]interface{}
}

type TargetDelegation struct {
	Keys  map[string]Key
	Roles []TargetDelegationRole
}

type TargetDelegationRole struct {
	Name             string
	KeyIDs           []string
	Threshold        int64
	PathHashPrefixes []string
	Paths            string
	Terminating      bool
}
