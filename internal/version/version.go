package version

import (
	"runtime"
	"time"
)

// SchemaVersion identifies the version command's JSON response schema.
const SchemaVersion = 1

var (
	Value   = "dev"
	Channel = "dev"
	Commit  = ""
	BuiltAt = ""
)

// Info is the version command's JSON response.
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

// Current returns metadata for the binary running this process.
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

// BuiltTime parses the optional build timestamp, returning zero when absent.
func BuiltTime() time.Time {
	value, err := time.Parse(time.RFC3339, BuiltAt)
	if err != nil {
		return time.Time{}
	}
	return value
}
