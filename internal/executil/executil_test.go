package executil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTermuxPathResolvesExecutableWithoutLookPath(t *testing.T) {
	prefix := t.TempDir()
	bin := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(bin, "pkg")
	if err := os.WriteFile(want, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	got, err := termuxPath("pkg", prefix)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolved path = %q, want %q", got, want)
	}
}

func TestTermuxPathReportsMissingCommand(t *testing.T) {
	_, err := termuxPath("pkg", t.TempDir())
	if err == nil {
		t.Fatal("missing command unexpectedly resolved")
	}
}
