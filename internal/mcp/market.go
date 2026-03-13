package mcp

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var marketLinePattern = regexp.MustCompile(`^\s*-\s+.*\*\*\[([^\]]+)\]\(([^)]+)\)\*\*\s+-\s+(.*)$`)

func SearchMarket(client *http.Client, keyword, userAgent string) ([]string, error) {
	request, err := http.NewRequest(http.MethodGet, "https://raw.githubusercontent.com/modelcontextprotocol/servers/main/README.md", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/plain")
	if strings.TrimSpace(userAgent) != "" {
		request.Header.Set("User-Agent", userAgent)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("market fetch: %s", response.Status)
	}

	keywordLower := strings.ToLower(strings.TrimSpace(keyword))
	results := make([]string, 0, 20)
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64<<10), 2<<20)
	for scanner.Scan() {
		line := scanner.Text()
		match := marketLinePattern.FindStringSubmatch(line)
		if len(match) != 4 {
			continue
		}
		name := strings.TrimSpace(match[1])
		link := strings.TrimSpace(match[2])
		desc := strings.TrimSpace(match[3])
		if name == "" || link == "" {
			continue
		}
		haystack := strings.ToLower(name + " " + desc)
		if !strings.Contains(haystack, keywordLower) {
			continue
		}
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			link = "https://github.com/modelcontextprotocol/servers/tree/main/" + strings.TrimPrefix(link, "./")
		}
		results = append(results, fmt.Sprintf("[market] %s - %s (%s)", name, desc, link))
		if len(results) >= 20 {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
