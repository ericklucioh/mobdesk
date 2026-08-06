package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBuildCommandPrefersExecutableWrappers(t *testing.T) {
	dir := t.TempDir()
	gradlew := filepath.Join(dir, "gradlew")
	if err := os.WriteFile(gradlew, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	mvnw := filepath.Join(dir, "mvnw")
	if err := os.WriteFile(mvnw, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got, err := ResolveBuildCommand(dir, "gradle"); err != nil || got != gradlew {
		t.Fatalf("gradle command = %q, %v", got, err)
	}
	if got, err := ResolveBuildCommand(dir, "maven"); err != nil || got != mvnw {
		t.Fatalf("maven command = %q, %v", got, err)
	}
}

func TestResolveBuildCommandFallsBackWithoutWrapper(t *testing.T) {
	dir := t.TempDir()
	for _, tool := range []struct{ name, want string }{{"gradle", "gradle"}, {"mvn", "mvn"}} {
		if got, err := ResolveBuildCommand(dir, tool.name); err != nil || got != tool.want {
			t.Fatalf("%s command = %q, %v", tool.name, got, err)
		}
	}
}

func TestResolveBuildCommandRejectsInvalidWrapper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gradlew")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveBuildCommand(dir, "gradle"); err == nil {
		t.Fatal("non-executable wrapper was accepted")
	}
}
