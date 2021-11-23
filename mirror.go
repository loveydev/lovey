package lovey

import "time"

type Mirrors struct {
	SpecVersion string
	Version     int64
	Expires     time.Time
	Mirrors     []Mirror
}

type Mirror struct {
	URLBase        string
	MetaPath       string
	TargetsPath    string
	MetaContent    []string
	TargetsContent []string
	Custom         map[string]interface{}
}
