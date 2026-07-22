package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSelectsStableAndStageChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		json.NewEncoder(response).Encode([]Release{
			{TagName: "v1.2.0"},
			{TagName: "test-v1.3.0", Prerelease: true},
		})
	}))
	defer server.Close()

	stable, err := Check(context.Background(), Options{APIBaseURL: server.URL, CurrentVersion: "v1.1.0", Channel: "stable", GOOS: "linux", GOARCH: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	if stable.LatestVersion != "v1.2.0" || !stable.Updated {
		t.Fatalf("unexpected stable result: %+v", stable)
	}
	stage, err := Check(context.Background(), Options{APIBaseURL: server.URL, CurrentVersion: "test-v1.3.0", Channel: "stage", GOOS: "linux", GOARCH: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	if stage.LatestVersion != "test-v1.3.0" || stage.Updated {
		t.Fatalf("unexpected stage result: %+v", stage)
	}
}

func TestTermuxNameserversReadsPrefixResolverConfig(t *testing.T) {
	prefix := t.TempDir()
	if err := os.MkdirAll(filepath.Join(prefix, "etc"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prefix, "etc", "resolv.conf"), []byte("nameserver 8.8.8.8\nnameserver 8.8.4.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PREFIX", prefix)

	nameservers := termuxNameservers()
	if len(nameservers) != 2 || nameservers[0] != "8.8.8.8" || nameservers[1] != "8.8.4.4" {
		t.Fatalf("nameservers = %v", nameservers)
	}
}

func TestApplyVerifiesChecksumAndReplacesBinary(t *testing.T) {
	content := []byte("new mobdesk binary")
	digest := sha256.Sum256(content)
	checksum := hex.EncodeToString(digest[:])
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/example/mobdesk/releases":
			json.NewEncoder(response).Encode([]Release{{
				TagName: "v1.1.0",
				Assets: []Asset{
					{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")},
					{Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")},
				},
			}})
		case "/download/mobdesk":
			response.Write(content)
		case "/download/checksums":
			fmt.Fprintf(response, "%s  mobdesk-linux-arm64\n", checksum)
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "mobdesk")
	if err := os.WriteFile(path, []byte("old mobdesk binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Apply(context.Background(), Options{
		APIBaseURL:     server.URL,
		Repository:     "example/mobdesk",
		CurrentVersion: "v1.0.0",
		Channel:        "stable",
		InstallPath:    path,
		GOOS:           "linux",
		GOARCH:         "arm64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Updated || result.LatestVersion != "v1.1.0" {
		t.Fatalf("unexpected update result: %+v", result)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != string(content) {
		t.Fatalf("binary content = %q, want %q", updated, content)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatalf("backup still exists: %v", err)
	}
}

func TestApplyRejectsInvalidChecksumWithoutReplacingBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/example/mobdesk/releases":
			json.NewEncoder(response).Encode([]Release{{
				TagName: "v1.1.0",
				Assets: []Asset{
					{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")},
					{Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")},
				},
			}})
		case "/download/mobdesk":
			response.Write([]byte("tampered"))
		case "/download/checksums":
			response.Write([]byte(strings.Repeat("0", sha256.Size*2) + "  mobdesk-linux-arm64\n"))
		}
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "mobdesk")
	if err := os.WriteFile(path, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Apply(context.Background(), Options{
		APIBaseURL:     server.URL,
		Repository:     "example/mobdesk",
		CurrentVersion: "v1.0.0",
		InstallPath:    path,
		GOOS:           "linux",
		GOARCH:         "arm64",
	})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("unexpected error: %v", err)
	}
	unchanged, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(unchanged) != "old" {
		t.Fatalf("binary changed after checksum failure: %q", unchanged)
	}
}

func serverURL(request *http.Request, path string) string {
	return "http://" + request.Host + path
}
