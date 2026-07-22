package update

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultRepository = "ericklucioh/mobdesk"

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
}

type Result struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Channel        string `json:"channel"`
	Asset          string `json:"asset"`
	InstallPath    string `json:"install_path,omitempty"`
	Updated        bool   `json:"updated"`
}

func Check(ctx context.Context, options Options) (Result, error) {
	options = options.withDefaults()
	if options.GOOS != "linux" || options.GOARCH != "arm64" {
		return Result{}, fmt.Errorf("release não disponível para %s/%s", options.GOOS, options.GOARCH)
	}
	release, err := latestRelease(ctx, options)
	if err != nil {
		return Result{}, err
	}
	assetName := options.BinaryName
	return Result{
		CurrentVersion: options.CurrentVersion,
		LatestVersion:  release.TagName,
		Channel:        options.Channel,
		Asset:          assetName,
		InstallPath:    options.InstallPath,
		Updated:        options.CurrentVersion != release.TagName,
	}, nil
}

func Apply(ctx context.Context, options Options) (Result, error) {
	options = options.withDefaults()
	result, err := Check(ctx, options)
	if err != nil {
		return Result{}, err
	}
	if !result.Updated {
		return result, nil
	}
	if options.InstallPath == "" {
		return result, fmt.Errorf("não foi possível detectar o caminho do executável")
	}
	if err := replaceBinary(ctx, options, result); err != nil {
		return result, err
	}
	result.Updated = true
	return result, nil
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
		o.BinaryName = fmt.Sprintf("mobdesk-%s-%s", o.GOOS, o.GOARCH)
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
	if len(nameservers) == 0 {
		return http.DefaultClient
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultClient
	}
	transport = transport.Clone()
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
	transport.DialContext = (&net.Dialer{Timeout: 30 * time.Second, Resolver: resolver}).DialContext
	return &http.Client{Transport: transport}
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
	defer file.Close()

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
		return Release{}, fmt.Errorf("consultar releases: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("consultar releases: HTTP %s", response.Status)
	}
	var releases []Release
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return Release{}, fmt.Errorf("ler releases: %w", err)
	}
	for _, release := range releases {
		if release.Draft || !matchesChannel(release, options.Channel) {
			continue
		}
		return release, nil
	}
	return Release{}, fmt.Errorf("nenhuma release disponível para o canal %s", options.Channel)
}

func matchesChannel(release Release, channel string) bool {
	if channel == "stage" {
		return release.Prerelease && strings.HasPrefix(release.TagName, "test-v")
	}
	return !release.Prerelease && strings.HasPrefix(release.TagName, "v")
}

func replaceBinary(ctx context.Context, options Options, result Result) error {
	if options.GOOS != "linux" || options.GOARCH != "arm64" {
		return fmt.Errorf("release não disponível para %s/%s", options.GOOS, options.GOARCH)
	}
	if _, err := os.Stat(options.InstallPath); err != nil {
		return fmt.Errorf("verificar executável atual: %w", err)
	}
	release, err := latestRelease(ctx, options)
	if err != nil {
		return err
	}
	binaryAsset, ok := findAsset(release, options.BinaryName)
	if !ok {
		return fmt.Errorf("release %s não possui o asset %s", release.TagName, options.BinaryName)
	}
	checksumAsset, ok := findAsset(release, options.ChecksumName)
	if !ok {
		return fmt.Errorf("release %s não possui o asset %s", release.TagName, options.ChecksumName)
	}
	expected, err := downloadChecksum(ctx, options.HTTPClient, checksumAsset.DownloadURL, options.BinaryName)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(options.InstallPath), ".mobdesk-update-*")
	if err != nil {
		return fmt.Errorf("criar arquivo temporário da atualização: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := downloadBinary(ctx, options.HTTPClient, binaryAsset.DownloadURL, temporary, expected); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o755); err != nil {
		temporary.Close()
		return fmt.Errorf("definir permissões do update: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("fechar arquivo temporário do update: %w", err)
	}

	backupPath := options.InstallPath + ".bak"
	_ = os.Remove(backupPath)
	if err := os.Rename(options.InstallPath, backupPath); err != nil {
		return fmt.Errorf("preparar substituição do executável: %w", err)
	}
	if err := os.Rename(temporaryPath, options.InstallPath); err != nil {
		_ = os.Rename(backupPath, options.InstallPath)
		return fmt.Errorf("substituir executável: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("remover backup do executável: %w", err)
	}
	return nil
}

func downloadChecksum(ctx context.Context, client HTTPClient, url, binaryName string) (string, error) {
	body, err := fetch(ctx, client, url)
	if err != nil {
		return "", fmt.Errorf("baixar checksums: %w", err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("ler checksums: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == binaryName && len(fields[0]) == sha256.Size*2 {
			if _, err := hex.DecodeString(fields[0]); err != nil {
				return "", fmt.Errorf("checksum inválido para %s: %w", binaryName, err)
			}
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("checksum não encontrado para %s", binaryName)
}

func downloadBinary(ctx context.Context, client HTTPClient, url string, destination io.Writer, expected string) error {
	body, err := fetch(ctx, client, url)
	if err != nil {
		return fmt.Errorf("baixar binário: %w", err)
	}
	defer body.Close()
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(destination, hash), body); err != nil {
		return fmt.Errorf("gravar binário: %w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum do binário não confere: esperado %s, obtido %s", expected, actual)
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
		response.Body.Close()
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
