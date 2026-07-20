package cobra

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupPhaseCanBeRecordedAndRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if setupPhaseDone("packages-installed") {
		t.Fatal("etapa não deveria existir antes do registro")
	}
	if err := markSetupPhase("packages-installed"); err != nil {
		t.Fatal(err)
	}
	if !setupPhaseDone("packages-installed") {
		t.Fatal("etapa registrada não foi encontrada")
	}
	info, err := os.Stat(filepath.Join(home, ".local", "share", "mobdesk", "state", "packages-installed.done"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissão inesperada: %o", info.Mode().Perm())
	}
}
