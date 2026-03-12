package targets

import "testing"

func TestSupportedTargets(t *testing.T) {
	if len(SupportedTargets) != 6 {
		t.Fatalf("expected 6 targets, got %d", len(SupportedTargets))
	}

	for _, name := range []string{"codex", "claude", "gemini", "qwen", "grok", "opencode"} {
		if _, ok := SupportedTargets[name]; !ok {
			t.Fatalf("missing target %s", name)
		}
	}
}

func TestValidateProfile(t *testing.T) {
	if err := ValidateProfile(DefaultProfile); err != nil {
		t.Fatalf("default profile should be valid: %v", err)
	}

	if err := ValidateProfile("unknown"); err == nil {
		t.Fatal("expected invalid profile to fail")
	}
}
