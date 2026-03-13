package skill

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var skillsDotShInstallPattern = regexp.MustCompile(`(?i)npx\s+skills\s+add\s+([^\s<]+)(?:\s+--skill\s+([^\s<]+))?`)

func isSkillsDotShURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	return strings.HasPrefix(raw, "https://skills.sh/") || raw == "https://skills.sh"
}

func ResolveSkillsDotSh(client *http.Client, pageURL, userAgent string) (string, string, error) {
	if !isSkillsDotShURL(pageURL) {
		return "", "", fmt.Errorf("not a skills.sh url: %s", pageURL)
	}
	request, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(userAgent) != "" {
		request.Header.Set("User-Agent", userAgent)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return "", "", fmt.Errorf("skills.sh: %s", response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return "", "", err
	}

	match := skillsDotShInstallPattern.FindSubmatch(body)
	if len(match) < 2 {
		return "", "", errors.New("skills.sh: install command not found")
	}
	repo := strings.TrimSpace(string(match[1]))
	skill := ""
	if len(match) > 2 {
		skill = strings.TrimSpace(string(match[2]))
	}
	if repo == "" {
		return "", "", errors.New("skills.sh: repository missing")
	}
	return repo, skill, nil
}
