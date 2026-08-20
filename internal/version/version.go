package version

import (
	"runtime"
	"time"
)

const SchemaVersion = 1

var (
	Value   = "dev"
	Channel = "dev"
	Commit  = ""
	BuiltAt = ""
)

type Info struct {
	SchemaVersion int    `json:"schema_version"`
	Command       string `json:"command"`
	Success       bool   `json:"success"`
	State         string `json:"state"`
	Message       string `json:"message"`
	Version       string `json:"version"`
	Channel       string `json:"channel"`
	Commit        string `json:"commit,omitempty"`
	BuiltAt       string `json:"built_at,omitempty"`
	GoVersion     string `json:"go_version"`
	OS            string `json:"os"`
	Architecture  string `json:"architecture"`
}

func Current() Info {
	return Info{
		SchemaVersion: SchemaVersion,
		Command:       "version",
		Success:       true,
		State:         "current",
		Version:       Value,
		Channel:       Channel,
		Commit:        Commit,
		BuiltAt:       BuiltAt,
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
	}
}

func BuiltTime() time.Time {
	value, err := time.Parse(time.RFC3339, BuiltAt)
	if err != nil {
		return time.Time{}
	}
	return value
}
