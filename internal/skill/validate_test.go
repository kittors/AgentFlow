package skill

import "testing"

func TestValidateSkillName(t *testing.T) {
	valid := []string{
		"agentflow",
		"turborepo",
		"next-upgrade",
		"foo_bar",
		"foo.bar",
		"foo-bar_1.2.3",
	}
	for _, name := range valid {
		if err := ValidateSkillName(name); err != nil {
			t.Fatalf("expected valid: %q: %v", name, err)
		}
	}

	invalid := []string{
		"",
		" ",
		".",
		"..",
		"../pwn",
		"..\\pwn",
		"pwn/child",
		"pwn\\child",
		"pwn..child",
		"pwn child",
		"pwn:child",
	}
	for _, name := range invalid {
		if err := ValidateSkillName(name); err == nil {
			t.Fatalf("expected invalid: %q", name)
		}
	}
}
