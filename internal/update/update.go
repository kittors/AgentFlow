package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/debuglog"
)

var (
	releaseTagAPI      = "https://api.github.com/repos/kittors/AgentFlow/releases/tags/continuous"
	releaseLatestAPI   = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"
	releaseAPIOverride = ""
	executablePath     = os.Executable
	evalSymlinks       = filepath.EvalSymlinks
)

const defaultHTTPTimeout = 2 * time.Minute

// ProgressFunc is called during SelfUpdateWithProgress to report current stage.
// stage is one of: "checking", "found", "downloading", "replacing".
// percent is 0–100 for download progress, or -1 for indeterminate stages.
// info carries optional contextual data (e.g. the new version string for "found").
type ProgressFunc func(stage string, percent int, info string)

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
	Name    string         `json:"name"`
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
		Client: &http.Client{Timeout: defaultHTTPTimeout},
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
	done := debuglog.Timed("update.Check")
	defer done()
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
			if c.Now().Unix()-entry.Timestamp < int64(ttl*3600) && isUsableCachedVersion(entry.Latest) {
				if !shouldRefreshMainBuildCache(current, entry.Latest) {
					result.Latest = entry.Latest
					result.UpdateAvailable = shouldUpdate(current, entry.Latest)
					return result, nil
				}
			}
		}
	}

	release, err := c.fetchLatestRelease()
	if err != nil {
		return result, err
	}
	result.Latest = releaseVersion(release)
	result.UpdateAvailable = shouldUpdate(current, result.Latest)
	_ = c.writeCache(cacheEntry{Latest: result.Latest, Timestamp: c.Now().Unix()})
	return result, nil
}

func (c *Checker) SelfUpdate(current string) (Result, error) {
	return c.SelfUpdateWithProgress(current, nil)
}

// SelfUpdateWithProgress behaves like SelfUpdate but calls progress at key
// stages so the caller (e.g. the TUI) can display live feedback.
func (c *Checker) SelfUpdateWithProgress(current string, progress ProgressFunc) (Result, error) {
	done := debuglog.Timed("update.SelfUpdateWithProgress")
	defer done()
	if progress == nil {
		progress = func(string, int, string) {}
	}

	current = normalizeVersion(current)
	progress("checking", -1, "")

	release, err := c.fetchLatestRelease()
	if err != nil {
		return Result{Current: current}, err
	}

	result := Result{
		Current: current,
		Latest:  releaseVersion(release),
	}
	result.UpdateAvailable = shouldUpdate(current, result.Latest)
	if !result.UpdateAvailable {
		return result, nil
	}

	progress("found", -1, result.Latest)

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

	progress("downloading", 0, result.Latest)
	if err := c.downloadAndReplace(downloadURL, executable, progress, result.Latest); err != nil {
		return result, err
	}

	_ = c.writeCache(cacheEntry{Latest: result.Latest, Timestamp: c.Now().Unix()})
	return result, nil
}

func (c *Checker) fetchLatestRelease() (releasePayload, error) {
	endpoints := []string{releaseTagAPI, releaseLatestAPI}
	if releaseAPIOverride != "" {
		endpoints = []string{releaseAPIOverride}
	}

	var lastErr error
	for _, endpoint := range endpoints {
		payload, err := c.fetchRelease(endpoint)
		if err == nil {
			return payload, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return releasePayload{}, lastErr
	}
	return releasePayload{}, errors.New("failed to resolve release metadata")
}

func (c *Checker) fetchRelease(endpoint string) (releasePayload, error) {
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
		return releasePayload{}, fmt.Errorf("%s: %s", endpoint, response.Status)
	}

	var payload releasePayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return releasePayload{}, err
	}
	return payload, nil
}

func releaseVersion(release releasePayload) string {
	if strings.TrimSpace(release.Name) != "" {
		if version := normalizeVersion(release.Name); version != "" && version != "unknown" {
			return version
		}
	}
	if strings.TrimSpace(release.TagName) != "" {
		if version := normalizeVersion(release.TagName); version != "" && version != "unknown" {
			return version
		}
	}
	return "unknown"
}

func (c *Checker) downloadAndReplace(downloadURL, destination string, progress ProgressFunc, version string) error {
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

	// Wrap the body with a progress reader when Content-Length is available.
	var body io.Reader = response.Body
	if progress != nil && response.ContentLength > 0 {
		body = &progressReader{
			reader: response.Body,
			total:  response.ContentLength,
			onProgress: func(percent int) {
				progress("downloading", percent, version)
			},
		}
	}

	tmpPath := destination + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, body); err != nil {
		file.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	progress("replacing", -1, version)

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

// progressReader wraps an io.Reader and calls onProgress with the current
// download percentage (0–100) whenever data is read.
type progressReader struct {
	reader     io.Reader
	total      int64
	read       int64
	lastReport int
	onProgress func(percent int)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.read += int64(n)
		percent := int(math.Min(float64(r.read)*100/float64(r.total), 100))
		// Only call back when the percentage actually changes to avoid flooding.
		if percent != r.lastReport {
			r.lastReport = percent
			if r.onProgress != nil {
				r.onProgress(percent)
			}
		}
	}
	return n, err
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
		if len(latestParts) > len(currentParts) {
			return true
		}
		if len(latestParts) < len(currentParts) {
			return false
		}
		if current == latest {
			return false
		}
		if strings.HasPrefix(current, "dev") {
			return false
		}
		return true
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
		limit := 0
		for limit < len(part) && part[limit] >= '0' && part[limit] <= '9' {
			limit++
		}
		if limit == 0 {
			return nil, false
		}
		value, err := strconv.Atoi(part[:limit])
		if err != nil {
			return nil, false
		}
		values = append(values, value)
		if limit != len(part) {
			break
		}
	}
	return values, true
}

func isUsableCachedVersion(version string) bool {
	version = normalizeVersion(version)
	if version == "" || version == "unknown" || version == "continuous" {
		return false
	}
	_, ok := parseReleaseVersion(version)
	return ok
}

func shouldRefreshMainBuildCache(current, cached string) bool {
	current = normalizeVersion(current)
	cached = normalizeVersion(cached)
	if current == "" || cached == "" || current == cached {
		return false
	}
	return strings.Contains(current, "-main.") && strings.Contains(cached, "-main.")
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
