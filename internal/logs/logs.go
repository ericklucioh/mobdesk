package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/status"
)

const (
	SchemaVersion = 1
	DefaultLines  = 20
)

type Options struct {
	Paths paths.Paths
	Name  string
	Lines int
}

type Record struct {
	Name          string    `json:"name"`
	Kind          string    `json:"kind"`
	State         string    `json:"state"`
	Version       string    `json:"version,omitempty"`
	LastAttemptAt time.Time `json:"last_attempt_at,omitempty"`
	InstalledAt   time.Time `json:"installed_at,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
	LogPath       string    `json:"log_path"`
	Content       string    `json:"content,omitempty"`
	Missing       bool      `json:"missing,omitempty"`
}

type Snapshot struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	Logs          []Record  `json:"logs"`
}

func Read(options Options) (Snapshot, error) {
	if options.Lines <= 0 {
		options.Lines = DefaultLines
	}

	directory := options.Paths.InstallationsDir()
	entries, err := os.ReadDir(directory)
	if err != nil && !os.IsNotExist(err) {
		return Snapshot{}, fmt.Errorf("ler registros de instalação: %w", err)
	}

	result := Snapshot{SchemaVersion: SchemaVersion, GeneratedAt: time.Now().UTC(), Logs: []Record{}}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		payload, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			continue
		}
		var installation status.InstallationStatus
		if json.Unmarshal(payload, &installation) != nil || installation.Name == "" {
			continue
		}
		if options.Name != "" && !strings.EqualFold(options.Name, installation.Name) {
			continue
		}
		record := Record{
			Name:          installation.Name,
			Kind:          installation.Kind,
			State:         installation.State,
			Version:       installation.Version,
			LastAttemptAt: installation.LastAttemptAt,
			InstalledAt:   installation.InstalledAt,
			LastError:     installation.LastError,
			LogPath:       installation.LogPath,
		}
		record.Content, record.Missing = readTail(record.LogPath, options.Lines)
		result.Logs = append(result.Logs, record)
	}
	sort.Slice(result.Logs, func(i, j int) bool {
		return result.Logs[i].Name < result.Logs[j].Name
	})
	return result, nil
}

func readTail(path string, lines int) (string, bool) {
	if path == "" {
		return "", true
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", true
	}
	text := strings.TrimRight(strings.ReplaceAll(string(payload), "\r\n", "\n"), "\n")
	if text == "" {
		return "", false
	}
	values := strings.Split(text, "\n")
	if len(values) > lines {
		values = values[len(values)-lines:]
	}
	return strings.TrimSpace(strings.Join(values, "\n")), false
}
