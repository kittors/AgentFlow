package skill

import (
	"fmt"
	"regexp"
	"strings"
)

var validSkillNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func ValidateSkillName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("invalid skill name: empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid skill name: %q", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid skill name: %q", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid skill name: %q", name)
	}
	if !validSkillNamePattern.MatchString(name) {
		return fmt.Errorf("invalid skill name: %q", name)
	}
	return nil
}
