package skill

import (
	"bufio"
	"os"
	"strings"
)

func ParseSkillName(path string) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			}
			break
		}
		if !inFrontMatter {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			value = strings.Trim(value, `"'`)
			if value != "" {
				return value, true
			}
		}
	}
	return "", false
}
