package cobra

import (
	"fmt"
	"os"
	"path/filepath"
)

func setupStateDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "mobdesk", "state")
}

func setupPhaseDone(phase string) bool {
	_, err := os.Stat(filepath.Join(setupStateDir(), phase+".done"))
	return err == nil
}

func markSetupPhase(phase string) error {
	stateDir := setupStateDir()
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return fmt.Errorf("criar estado da etapa %s: %w", phase, err)
	}
	path := filepath.Join(stateDir, phase+".done")
	if err := os.WriteFile(path, []byte("concluida\n"), 0o600); err != nil {
		return fmt.Errorf("registrar etapa %s: %w", phase, err)
	}
	return nil
}
