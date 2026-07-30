package update

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

const (
	DefaultRepository  = "ericklucioh/mobdesk"
	maxReleaseResponse = 1 << 20
	maxChecksumSize    = 1 << 20
	maxBinarySize      = 128 << 20
	selfTestTimeout    = 10 * time.Second
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	Assets     []Asset `json:"assets"`
}

type Options struct {
	HTTPClient     HTTPClient
	Repository     string
	CurrentVersion string
	Channel        string
	InstallPath    string
	GOOS           string
	GOARCH         string
	BinaryName     string
	ChecksumName   string
	APIBaseURL     string
	ValidateBinary func(context.Context, string, string) error
}

type Result struct {
	SchemaVersion  int    `json:"schema_version"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Channel        string `json:"channel"`
	Asset          string `json:"asset"`
	InstallPath    string `json:"install_path,omitempty"`
	Updated        bool   `json:"updated"`
}

func Check(ctx context.Context, options Options) (result Result, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = updateError(err)
		}
	}()
	options = options.withDefaults()
	if !supportsReleaseTarget(options.GOOS, options.GOARCH) {
		return Result{}, fmt.Errorf("release unavailable for %s/%s", options.GOOS, options.GOARCH)
	}
	release, err := latestRelease(ctx, options)
	if err != nil {
		return Result{}, err
	}
	assetName := options.BinaryName
	result = Result{
		SchemaVersion:  1,
		CurrentVersion: options.CurrentVersion,
		LatestVersion:  release.TagName,
		Channel:        options.Channel,
		Asset:          assetName,
		InstallPath:    options.InstallPath,
		Updated:        options.CurrentVersion != release.TagName,
	}
	return result, nil
}

func Apply(ctx context.Context, options Options) (result Result, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = updateError(err)
		}
	}()
	options = options.withDefaults()
	if err := recoverInterruptedUpdate(options.InstallPath); err != nil {
		return Result{}, err
	}
	result, err = Check(ctx, options)
	if err != nil {
		return Result{}, err
	}
	if !result.Updated {
		return result, nil
	}
	if options.InstallPath == "" {
		return result, fmt.Errorf("could not detect executable path")
	}
	if err := replaceBinary(ctx, options, result); err != nil {
		return result, err
	}
	result.Updated = true
	return result, nil
}

// Recover restores the last working binary when an interrupted update left
// only its backup. It is safe to call before checking for updates.
func Recover(options Options) error {
	err := recoverInterruptedUpdate(options.withDefaults().InstallPath)
	if err != nil && i18n.ErrorCode(err) == "" {
		return updateError(err)
	}
	return err
}

func updateError(cause error) error {
	return i18n.NewError(i18n.ServiceUpdateError, "update_operation_failed", map[string]any{"Detail": cause.Error()}, cause)
}

func (o Options) withDefaults() Options {
	if o.HTTPClient == nil {
		o.HTTPClient = defaultHTTPClient()
	}
	if o.Repository == "" {
		o.Repository = DefaultRepository
	}
	if o.Channel == "" {
		o.Channel = "stable"
	}
	if o.GOOS == "" {
		o.GOOS = runtime.GOOS
	}
	if o.GOARCH == "" {
		o.GOARCH = runtime.GOARCH
	}
	if o.BinaryName == "" {
		assetGOOS := o.GOOS
		// Termux reports android, but releases are static Linux ARM64 binaries
		// and the release workflow uses that target explicitly.
		if assetGOOS == "android" {
			assetGOOS = "linux"
		}
		o.BinaryName = fmt.Sprintf("mobdesk-%s-%s", assetGOOS, o.GOARCH)
	}
	if o.ChecksumName == "" {
		o.ChecksumName = "SHA256SUMS"
	}
	if o.APIBaseURL == "" {
		o.APIBaseURL = "https://api.github.com"
	}
	if o.InstallPath == "" {
		o.InstallPath = executablePath()
	}
	return o
}

func defaultHTTPClient() HTTPClient {
	nameservers := termuxNameservers()
	rootCAs := termuxRootCAs()
	if len(nameservers) == 0 && rootCAs == nil {
		return &http.Client{Timeout: 45 * time.Second}
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: 45 * time.Second}
	}
	transport = transport.Clone()
	if rootCAs != nil {
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		if transport.TLSClientConfig != nil {
			tlsConfig = transport.TLSClientConfig.Clone()
		}
		tlsConfig.RootCAs = rootCAs
		transport.TLSClientConfig = tlsConfig
	}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var lastErr error
			for _, nameserver := range nameservers {
				connection, err := (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "udp", net.JoinHostPort(nameserver, "53"))
				if err == nil {
					return connection, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
	}
	if len(nameservers) > 0 {
		transport.DialContext = (&net.Dialer{Timeout: 30 * time.Second, Resolver: resolver}).DialContext
	}
	return &http.Client{Transport: transport, Timeout: 45 * time.Second}
}

func termuxNameservers() []string {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		return nil
	}
	file, err := os.Open(filepath.Join(prefix, "etc", "resolv.conf"))
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "nameserver" && net.ParseIP(fields[1]) != nil {
			result = append(result, fields[1])
		}
	}
	return result
}

func termuxRootCAs() *x509.CertPool {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(prefix, "etc", "tls", "cert.pem"))
	if err != nil {
		return nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(data) {
		return nil
	}
	return pool
}

func latestRelease(ctx context.Context, options Options) (Release, error) {
	url := strings.TrimRight(options.APIBaseURL, "/") + "/repos/" + options.Repository + "/releases?per_page=100"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "mobdesk-updater")
	response, err := options.HTTPClient.Do(request)
	if err != nil {
		return Release{}, fmt.Errorf("query releases: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("query releases: HTTP %s", response.Status)
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, maxReleaseResponse+1))
	if err != nil {
		return Release{}, fmt.Errorf("read releases: %w", err)
	}
	if len(payload) > maxReleaseResponse {
		return Release{}, fmt.Errorf("release response exceeds the allowed limit")
	}
	var releases []Release
	if err := json.Unmarshal(payload, &releases); err != nil {
		return Release{}, fmt.Errorf("read releases: %w", err)
	}
	for _, release := range releases {
		if release.Draft || !matchesChannel(release, options.Channel) {
			continue
		}
		return release, nil
	}
	return Release{}, fmt.Errorf("no release available for channel %s", options.Channel)
}

func matchesChannel(release Release, channel string) bool {
	if channel == "stage" {
		return release.Prerelease && strings.HasPrefix(release.TagName, "test-v")
	}
	return !release.Prerelease && strings.HasPrefix(release.TagName, "v")
}

func replaceBinary(ctx context.Context, options Options, result Result) error {
	if !supportsReleaseTarget(options.GOOS, options.GOARCH) {
		return fmt.Errorf("release unavailable for %s/%s", options.GOOS, options.GOARCH)
	}
	if _, err := os.Stat(options.InstallPath); err != nil {
		return fmt.Errorf("check current executable: %w", err)
	}
	release, err := latestRelease(ctx, options)
	if err != nil {
		return err
	}
	binaryAsset, ok := findAsset(release, options.BinaryName)
	if !ok {
		return fmt.Errorf("release %s does not contain asset %s", release.TagName, options.BinaryName)
	}
	checksumAsset, ok := findAsset(release, options.ChecksumName)
	if !ok {
		return fmt.Errorf("release %s does not contain asset %s", release.TagName, options.ChecksumName)
	}
	expected, err := downloadChecksum(ctx, options.HTTPClient, checksumAsset.DownloadURL, options.BinaryName)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(options.InstallPath), ".mobdesk-update-*")
	if err != nil {
		return fmt.Errorf("create update temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := downloadBinary(ctx, options.HTTPClient, binaryAsset.DownloadURL, temporary, expected); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o755); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set update permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close update temporary file: %w", err)
	}
	if err := validateBinary(ctx, options, temporaryPath, release.TagName); err != nil {
		return fmt.Errorf("validate downloaded binary: %w", err)
	}

	backupPath := options.InstallPath + ".bak"
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous backup: %w", err)
	}
	if err := os.Rename(options.InstallPath, backupPath); err != nil {
		return fmt.Errorf("create executable backup: %w", err)
	}
	if err := os.Rename(temporaryPath, options.InstallPath); err != nil {
		if restoreErr := os.Rename(backupPath, options.InstallPath); restoreErr != nil {
			return fmt.Errorf("activate update: %w; restore backup: %v", err, restoreErr)
		}
		return fmt.Errorf("activate update: %w", err)
	}
	if err := validateBinary(ctx, options, options.InstallPath, release.TagName); err != nil {
		_ = os.Remove(options.InstallPath)
		if restoreErr := os.Rename(backupPath, options.InstallPath); restoreErr != nil {
			return fmt.Errorf("validate active update: %w; restore backup: %v", err, restoreErr)
		}
		return fmt.Errorf("validate active update: %w; previous version restored", err)
	}
	return nil
}

func recoverInterruptedUpdate(installPath string) error {
	if installPath == "" {
		return nil
	}
	if _, err := os.Stat(installPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check current executable: %w", err)
	}

	backupPath := installPath + ".bak"
	if _, err := os.Stat(backupPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("executable and backup not found; reinstall with go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest")
		}
		return fmt.Errorf("check update backup: %w", err)
	}
	if err := os.Rename(backupPath, installPath); err != nil {
		return fmt.Errorf("recover interrupted update: %w", err)
	}
	return nil
}

func supportsReleaseTarget(goos, goarch string) bool {
	return goarch == "arm64" && (goos == "linux" || goos == "android")
}

func downloadChecksum(ctx context.Context, client HTTPClient, url, binaryName string) (string, error) {
	body, err := fetch(ctx, client, url)
	if err != nil {
		return "", fmt.Errorf("download checksums: %w", err)
	}
	defer func() { _ = body.Close() }()
	data, err := io.ReadAll(io.LimitReader(body, maxChecksumSize+1))
	if err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	if len(data) > maxChecksumSize {
		return "", fmt.Errorf("checksums exceed the allowed limit")
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == binaryName && len(fields[0]) == sha256.Size*2 {
			if _, err := hex.DecodeString(fields[0]); err != nil {
				return "", fmt.Errorf("invalid checksum for %s: %w", binaryName, err)
			}
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", binaryName)
}

func downloadBinary(ctx context.Context, client HTTPClient, url string, destination io.Writer, expected string) error {
	body, err := fetch(ctx, client, url)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	defer func() { _ = body.Close() }()
	hash := sha256.New()
	bytesWritten, err := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(body, maxBinarySize+1))
	if err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	if bytesWritten == 0 {
		return fmt.Errorf("binary download is empty")
	}
	if bytesWritten > maxBinarySize {
		return fmt.Errorf("binary download exceeds the allowed limit")
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("binary checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func validateBinary(ctx context.Context, options Options, path, expectedVersion string) error {
	if options.ValidateBinary != nil {
		return options.ValidateBinary(ctx, path, expectedVersion)
	}
	testContext, cancel := context.WithTimeout(ctx, selfTestTimeout)
	defer cancel()
	command := exec.CommandContext(testContext, path, "version", "--json")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("run version --json: %w", err)
	}
	var info struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output, &info); err != nil {
		return fmt.Errorf("read version --json: %w", err)
	}
	if info.Version != expectedVersion {
		return fmt.Errorf("binary version is %q, expected %q", info.Version, expectedVersion)
	}
	return nil
}

func fetch(ctx context.Context, client HTTPClient, url string) (io.ReadCloser, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "mobdesk-updater")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("HTTP %s", response.Status)
	}
	return response.Body, nil
}

func findAsset(release Release, name string) (Asset, bool) {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return Asset{}, false
}

func executablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}
	return path
}
