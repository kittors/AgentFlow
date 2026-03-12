package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const releaseAPI = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"

var (
	releaseAPIOverride = ""
	executablePath     = os.Executable
	evalSymlinks       = filepath.EvalSymlinks
)

type Options struct {
	Force         bool
	CacheTTLHours int
}

type Result struct {
	Current         string
	Latest          string
	UpdateAvailable bool
}

type cacheEntry struct {
	Latest    string `json:"latest"`
	Timestamp int64  `json:"timestamp"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type releasePayload struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type Checker struct {
	Client    *http.Client
	Now       func() time.Time
	CacheFile string
}

func NewChecker() *Checker {
	homeDir, _ := os.UserHomeDir()
	return &Checker{
		Client: &http.Client{Timeout: 5 * time.Second},
		Now:    time.Now,
		CacheFile: filepath.Join(
			homeDir,
			".cache",
			"agentflow",
			"version_cache.json",
		),
	}
}

func (c *Checker) Check(current string, options Options) (Result, error) {
	current = normalizeVersion(current)
	result := Result{Current: current}

	ttl := options.CacheTTLHours
	if ttl == 0 {
		return result, nil
	}
	if ttl < 0 {
		ttl = 72
	}

	if !options.Force {
		if entry, err := c.readCache(); err == nil && entry != nil {
			if c.Now().Unix()-entry.Timestamp < int64(ttl*3600) {
				result.Latest = entry.Latest
				result.UpdateAvailable = shouldUpdate(current, entry.Latest)
				return result, nil
			}
		}
	}

	release, err := c.fetchLatestRelease()
	if err != nil {
		return result, err
	}
	result.Latest = normalizeVersion(release.TagName)
	result.UpdateAvailable = shouldUpdate(current, result.Latest)
	_ = c.writeCache(cacheEntry{Latest: result.Latest, Timestamp: c.Now().Unix()})
	return result, nil
}

func (c *Checker) SelfUpdate(current string) (Result, error) {
	current = normalizeVersion(current)
	release, err := c.fetchLatestRelease()
	if err != nil {
		return Result{Current: current}, err
	}

	result := Result{
		Current: current,
		Latest:  normalizeVersion(release.TagName),
	}
	result.UpdateAvailable = shouldUpdate(current, result.Latest)
	if !result.UpdateAvailable {
		return result, nil
	}

	assetName, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return result, err
	}
	downloadURL := ""
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return result, fmt.Errorf("release asset not found: %s", assetName)
	}

	executable, err := executablePath()
	if err != nil {
		return result, err
	}
	if resolved, err := evalSymlinks(executable); err == nil {
		executable = resolved
	}

	if runtime.GOOS == "windows" {
		return result, errors.New("self-update on Windows is not supported yet; rerun install.ps1")
	}
	if err := c.downloadAndReplace(downloadURL, executable); err != nil {
		return result, err
	}

	_ = c.writeCache(cacheEntry{Latest: result.Latest, Timestamp: c.Now().Unix()})
	return result, nil
}

func (c *Checker) fetchLatestRelease() (releasePayload, error) {
	endpoint := releaseAPI
	if releaseAPIOverride != "" {
		endpoint = releaseAPIOverride
	}

	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return releasePayload{}, err
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("User-Agent", "agentflow-go")

	response, err := c.Client.Do(request)
	if err != nil {
		return releasePayload{}, err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return releasePayload{}, errors.New(response.Status)
	}

	var payload releasePayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return releasePayload{}, err
	}
	return payload, nil
}

func (c *Checker) downloadAndReplace(downloadURL, destination string) error {
	request, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "agentflow-go")

	response, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return errors.New(response.Status)
	}

	tmpPath := destination + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	mode := os.FileMode(0o755)
	if info, err := os.Stat(destination); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) {
		return os.Rename(tmpPath, destination)
	}

	backupPath := destination + ".bak"
	_ = os.Remove(backupPath)
	if err := os.Rename(destination, backupPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, destination); err != nil {
		_ = os.Rename(backupPath, destination)
		_ = os.Remove(tmpPath)
		return err
	}
	_ = os.Remove(backupPath)
	return nil
}

func assetNameForPlatform(goos, goarch string) (string, error) {
	switch goos {
	case "linux", "darwin", "windows":
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	switch goarch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	suffix := ""
	if goos == "windows" {
		suffix = ".exe"
	}
	return fmt.Sprintf("agentflow-%s-%s%s", goos, goarch, suffix), nil
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	if version == "" {
		return "unknown"
	}
	return version
}

func shouldUpdate(current, latest string) bool {
	if latest == "" || latest == "unknown" {
		return false
	}
	if current == "unknown" {
		return true
	}

	currentParts, currentOK := parseReleaseVersion(current)
	latestParts, latestOK := parseReleaseVersion(latest)
	if currentOK && latestOK {
		for index := 0; index < len(currentParts) && index < len(latestParts); index++ {
			if latestParts[index] > currentParts[index] {
				return true
			}
			if latestParts[index] < currentParts[index] {
				return false
			}
		}
		return len(latestParts) > len(currentParts)
	}

	if strings.HasPrefix(current, "dev") {
		return false
	}
	return latest != current
}

func parseReleaseVersion(version string) ([]int, bool) {
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return nil, false
	}

	values := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		values = append(values, value)
	}
	return values, true
}

func (c *Checker) readCache() (*cacheEntry, error) {
	data, err := os.ReadFile(c.CacheFile)
	if err != nil {
		return nil, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (c *Checker) writeCache(entry cacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(c.CacheFile), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(c.CacheFile, data, 0o644)
}
