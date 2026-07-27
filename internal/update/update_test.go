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
		_ = json.NewEncoder(response).Encode([]Release{
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

func TestTermuxUsesLinuxARM64ReleaseAsset(t *testing.T) {
	options := (Options{GOOS: "android", GOARCH: "arm64"}).withDefaults()
	if !supportsReleaseTarget(options.GOOS, options.GOARCH) {
		t.Fatal("android/arm64 Termux target should use the Linux ARM64 release")
	}
	if options.BinaryName != "mobdesk-linux-arm64" {
		t.Fatalf("binary asset = %q, want mobdesk-linux-arm64", options.BinaryName)
	}
}

func TestUnsupportedReleaseTarget(t *testing.T) {
	if supportsReleaseTarget("android", "amd64") || supportsReleaseTarget("darwin", "arm64") {
		t.Fatal("unsupported release target was accepted")
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
			_ = json.NewEncoder(response).Encode([]Release{{
				TagName: "v1.1.0",
				Assets: []Asset{
					{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")},
					{Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")},
				},
			}})
		case "/download/mobdesk":
			_, _ = response.Write(content)
		case "/download/checksums":
			_, _ = fmt.Fprintf(response, "%s  mobdesk-linux-arm64\n", checksum)
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
		ValidateBinary: validBinary,
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
	backup, err := os.ReadFile(path + ".bak")
	if err != nil || string(backup) != "old mobdesk binary" {
		t.Fatalf("backup = %q, err = %v", backup, err)
	}
}

func TestApplyRestoresBackupWhenActiveBinaryFailsValidation(t *testing.T) {
	content := []byte("new mobdesk binary")
	digest := sha256.Sum256(content)
	checksum := hex.EncodeToString(digest[:])
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/example/mobdesk/releases":
			_ = json.NewEncoder(response).Encode([]Release{{TagName: "v1.1.0", Assets: []Asset{{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")}, {Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")}}}})
		case "/download/mobdesk":
			_, _ = response.Write(content)
		case "/download/checksums":
			_, _ = fmt.Fprintf(response, "%s  mobdesk-linux-arm64\n", checksum)
		}
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "mobdesk")
	if err := os.WriteFile(path, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	validations := 0
	_, err := Apply(context.Background(), Options{APIBaseURL: server.URL, Repository: "example/mobdesk", CurrentVersion: "v1.0.0", InstallPath: path, GOOS: "linux", GOARCH: "arm64", ValidateBinary: func(_ context.Context, _ string, _ string) error {
		validations++
		if validations == 2 {
			return fmt.Errorf("binary cannot start")
		}
		return nil
	}})
	if err == nil || !strings.Contains(err.Error(), "versão anterior restaurada") {
		t.Fatalf("unexpected error: %v", err)
	}
	active, readErr := os.ReadFile(path)
	if readErr != nil || string(active) != "old" {
		t.Fatalf("active = %q, err = %v", active, readErr)
	}
}

func TestApplyRecoversBackupFromInterruptedUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobdesk")
	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, []byte("previous mobdesk binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/example/mobdesk/releases" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(response).Encode([]Release{{TagName: "v1.0.0"}}); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

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
	if result.Updated {
		t.Fatalf("update unexpectedly applied: %+v", result)
	}
	recovered, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(recovered) != "previous mobdesk binary" {
		t.Fatalf("recovered binary = %q", recovered)
	}
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Fatalf("backup remains after recovery: %v", err)
	}
}

func TestRecoverInterruptedUpdateLeavesExistingBinaryUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mobdesk")
	if err := os.WriteFile(path, []byte("current"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".bak", []byte("stale backup"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := recoverInterruptedUpdate(path); err != nil {
		t.Fatal(err)
	}
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "current" {
		t.Fatalf("current binary = %q", current)
	}
}

func TestApplyRejectsInvalidChecksumWithoutReplacingBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/example/mobdesk/releases":
			_ = json.NewEncoder(response).Encode([]Release{{
				TagName: "v1.1.0",
				Assets: []Asset{
					{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")},
					{Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")},
				},
			}})
		case "/download/mobdesk":
			_, _ = response.Write([]byte("tampered"))
		case "/download/checksums":
			_, _ = response.Write([]byte(strings.Repeat("0", sha256.Size*2) + "  mobdesk-linux-arm64\n"))
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

func TestApplyRejectsEmptyBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/example/mobdesk/releases":
			_ = json.NewEncoder(response).Encode([]Release{{
				TagName: "v1.1.0",
				Assets: []Asset{
					{Name: "mobdesk-linux-arm64", DownloadURL: serverURL(request, "/download/mobdesk")},
					{Name: "SHA256SUMS", DownloadURL: serverURL(request, "/download/checksums")},
				},
			}})
		case "/download/mobdesk":
			// O checksum corresponde ao conteúdo vazio; mesmo assim o updater
			// deve recusar substituir um executável válido por zero bytes.
		case "/download/checksums":
			_, _ = fmt.Fprintf(response, "%s  mobdesk-linux-arm64\n", strings.Repeat("0", sha256.Size*2))
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
	if err == nil || !strings.Contains(err.Error(), "download do binário vazio") {
		t.Fatalf("unexpected error: %v", err)
	}
	unchanged, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(unchanged) != "old" {
		t.Fatalf("binary changed after empty download: %q", unchanged)
	}
}

func TestDownloadChecksumRejectsAmbiguousManifestLine(t *testing.T) {
	digest := strings.Repeat("0", sha256.Size*2)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(response, "%s  mobdesk-linux-arm64 extra-field\n", digest)
	}))
	defer server.Close()

	_, err := downloadChecksum(context.Background(), server.Client(), server.URL, "mobdesk-linux-arm64")
	if err == nil || !strings.Contains(err.Error(), "não encontrado") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecoverInterruptedUpdateErrorsWithoutBinaryOrBackup(t *testing.T) {
	err := recoverInterruptedUpdate(filepath.Join(t.TempDir(), "mobdesk"))
	if err == nil || !strings.Contains(err.Error(), "go install") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func serverURL(request *http.Request, path string) string {
	return "http://" + request.Host + path
}

func validBinary(context.Context, string, string) error { return nil }
