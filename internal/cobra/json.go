package cobra

import "os"

type operationResult struct {
	SchemaVersion int      `json:"schema_version"`
	Command       string   `json:"command"`
	Success       bool     `json:"success"`
	State         string   `json:"state"`
	Message       string   `json:"message"`
	Port          int      `json:"port,omitempty"`
	Addresses     []string `json:"addresses,omitempty"`
}

func withQuietStdout(run func() error) error {
	previous := os.Stdout
	quiet, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	os.Stdout = quiet
	runErr := run()
	os.Stdout = previous
	_ = quiet.Close()
	return runErr
}
