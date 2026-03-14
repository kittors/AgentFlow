package skill

import (
	"bufio"
	"os"
	"strings"
)

type FrontMatter struct {
	Name        string
	Description string
}

func ParseSkillName(path string) (string, bool) {
	meta, ok := ParseSkillFrontMatter(path)
	if !ok || strings.TrimSpace(meta.Name) == "" {
		return "", false
	}
	return meta.Name, true
}

func ParseSkillDescription(path string) (string, bool) {
	meta, ok := ParseSkillFrontMatter(path)
	if !ok || strings.TrimSpace(meta.Description) == "" {
		return "", false
	}
	return meta.Description, true
}

func ParseSkillFrontMatter(path string) (FrontMatter, bool) {
	file, err := os.Open(path)
	if err != nil {
		return FrontMatter{}, false
	}
	defer file.Close()

	meta := FrontMatter{}
	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	pending := ""

	for {
		raw := pending
		if raw != "" {
			pending = ""
		} else {
			if !scanner.Scan() {
				break
			}
			raw = scanner.Text()
		}

		line := strings.TrimSpace(raw)
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			}
			// End of front matter.
			return meta, true
		}
		if !inFrontMatter {
			continue
		}

		switch {
		case strings.HasPrefix(line, "name:"):
			value := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			value = strings.Trim(value, `"'`)
			if value != "" {
				meta.Name = value
			}
		case strings.HasPrefix(line, "description:"):
			value := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			value = strings.Trim(value, `"'`)
			if value == "|" || value == ">" || value == "" {
				lines := []string{}
				for scanner.Scan() {
					blockRaw := scanner.Text()
					blockTrim := strings.TrimSpace(blockRaw)
					if blockTrim == "---" {
						meta.Description = strings.TrimSpace(strings.Join(lines, "\n"))
						return meta, true
					}
					if blockRaw == "" || strings.HasPrefix(blockRaw, " ") || strings.HasPrefix(blockRaw, "\t") {
						lines = append(lines, strings.TrimSpace(blockRaw))
						continue
					}
					// Non-indented line means we've exited the block; process it next.
					pending = blockRaw
					break
				}
				meta.Description = strings.TrimSpace(strings.Join(lines, "\n"))
				continue
			}
			if value != "" {
				meta.Description = value
			}
		}
	}

	if !inFrontMatter {
		return meta, false
	}
	return meta, true
}
