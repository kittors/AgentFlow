package skill

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	githubRepoPattern = regexp.MustCompile(`(?i)^(?:https://github\.com/)?([a-z0-9_.-]+)/([a-z0-9_.-]+?)(?:\.git)?/?$`)
)

type githubRepoPayload struct {
	DefaultBranch string `json:"default_branch"`
}

func ParseGitHubRepo(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("missing repository")
	}

	if strings.HasPrefix(raw, "http://") {
		return "", "", errors.New("only https repositories are supported")
	}
	if strings.HasPrefix(raw, "https://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", "", err
		}
		if !strings.EqualFold(parsed.Host, "github.com") {
			return "", "", fmt.Errorf("unsupported host: %s", parsed.Host)
		}
		raw = strings.TrimPrefix(parsed.Path, "/")
		raw = strings.TrimSuffix(raw, "/")
	}

	match := githubRepoPattern.FindStringSubmatch(raw)
	if len(match) != 3 {
		return "", "", fmt.Errorf("invalid GitHub repository: %s", raw)
	}
	return match[1], match[2], nil
}

func ResolveDefaultBranch(client *http.Client, owner, repo, userAgent string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	if strings.TrimSpace(userAgent) != "" {
		request.Header.Set("User-Agent", userAgent)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("github api: %s", response.Status)
	}

	var payload githubRepoPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.DefaultBranch) == "" {
		return "", errors.New("github api: default_branch missing")
	}
	return payload.DefaultBranch, nil
}

func DownloadGitHubZip(client *http.Client, cacheDir, owner, repo, ref, userAgent string) (string, error) {
	downloadURL := fmt.Sprintf("https://codeload.github.com/%s/%s/zip/%s", owner, repo, ref)

	request, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(userAgent) != "" {
		request.Header.Set("User-Agent", userAgent)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download zip: %s", response.Status)
	}

	tmp, err := os.CreateTemp(cacheDir, "repo-*.zip")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := io.Copy(tmp, response.Body); err != nil {
		cleanup()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}
