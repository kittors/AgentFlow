package targets

import "testing"

func TestSupportedTargets(t *testing.T) {
	if len(SupportedTargets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(SupportedTargets))
	}

	for _, name := range []string{"codex", "claude"} {
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
