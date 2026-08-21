package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveBuildCommand selects a project wrapper without ever modifying it.
// The returned command path is absolute when a wrapper exists.
// ResolveBuildCommand selects a project-local build wrapper when one exists.
func ResolveBuildCommand(projectDir, tool string) (string, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return "", fmt.Errorf("project directory is required")
	}

	var wrapper, fallback string
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "gradle":
		wrapper, fallback = "gradlew", "gradle"
	case "maven", "mvn":
		wrapper, fallback = "mvnw", "mvn"
	default:
		return "", fmt.Errorf("unsupported build tool %q", tool)
	}

	path := filepath.Join(projectDir, wrapper)
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() || info.Mode()&0o111 == 0 {
			return "", fmt.Errorf("build wrapper %q exists but is not executable", path)
		}
		return path, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect build wrapper %q: %w", path, err)
	}
	return fallback, nil
}
