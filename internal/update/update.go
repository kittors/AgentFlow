package update

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const releaseAPI = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"

var releaseAPIOverride = ""

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
	if current == "" {
		current = "unknown"
	}
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
				result.UpdateAvailable = entry.Latest != "" && entry.Latest != current
				return result, nil
			}
		}
	}

	latest, err := c.fetchLatest()
	if err != nil {
		return result, err
	}
	result.Latest = latest
	result.UpdateAvailable = latest != "" && latest != current
	_ = c.writeCache(cacheEntry{Latest: latest, Timestamp: c.Now().Unix()})
	return result, nil
}

func (c *Checker) fetchLatest() (string, error) {
	endpoint := releaseAPI
	if releaseAPIOverride != "" {
		endpoint = releaseAPIOverride
	}
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("User-Agent", "agentflow-go")

	response, err := c.Client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return "", errors.New(response.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimPrefix(payload.TagName, "v"), nil
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
