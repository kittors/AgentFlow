package skill

import "testing"

func TestSkillsDotShInstallPattern(t *testing.T) {
	html := []byte(`<code>$ npx skills add https://github.com/vercel/turborepo --skill turborepo</code>`)
	match := skillsDotShInstallPattern.FindSubmatch(html)
	if len(match) < 2 {
		t.Fatalf("expected match")
	}
	if got := string(match[1]); got != "https://github.com/vercel/turborepo" {
		t.Fatalf("repo mismatch: %q", got)
	}
	if got := string(match[2]); got != "turborepo" {
		t.Fatalf("skill mismatch: %q", got)
	}
}
