package workstation

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractWorkspaceTarCopiesRegularFiles(t *testing.T) {
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	if err := writer.WriteHeader(&tar.Header{Name: "project", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	contents := []byte("package main\n")
	if err := writer.WriteHeader(&tar.Header{Name: "project/main.go", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err := extractWorkspaceTar(&buffer, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(destination, "project", "main.go"))
	if err != nil || string(got) != string(contents) {
		t.Fatalf("migrated file = %q, err = %v", got, err)
	}
}

func TestExtractWorkspaceTarRejectsUnsafeLinksAndTraversal(t *testing.T) {
	for _, header := range []*tar.Header{
		{Name: "../outside", Typeflag: tar.TypeReg, Size: 0},
		{Name: "project/link", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"},
		{Name: "project/link", Typeflag: tar.TypeSymlink, Linkname: "../../outside"},
	} {
		var buffer bytes.Buffer
		writer := tar.NewWriter(&buffer)
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := extractWorkspaceTar(&buffer, t.TempDir()); err == nil {
			t.Fatalf("header %+v was accepted", header)
		}
	}
}

func TestExtractWorkspaceTarPreservesInternalSymlink(t *testing.T) {
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	if err := writer.WriteHeader(&tar.Header{Name: "project/link", Typeflag: tar.TypeSymlink, Linkname: "../shared"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err := extractWorkspaceTar(&buffer, destination); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(filepath.Join(destination, "project", "link"))
	if err != nil || target != "../shared" {
		t.Fatalf("migrated link = %q, err = %v", target, err)
	}
}
