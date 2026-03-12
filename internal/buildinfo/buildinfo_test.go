package buildinfo

import (
	"runtime/debug"
	"testing"
)

func TestCurrentVersionPrefersInjectedRelease(t *testing.T) {
	originalVersion := Version
	originalReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		Version = originalVersion
		readBuildInfo = originalReadBuildInfo
	})

	Version = "1.2.3"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}

	if got := CurrentVersion(); got != "1.2.3" {
		t.Fatalf("expected injected version, got %q", got)
	}
}

func TestCurrentVersionFallsBackToVCSRevision(t *testing.T) {
	originalVersion := Version
	originalReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		Version = originalVersion
		readBuildInfo = originalReadBuildInfo
	})

	Version = "dev"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "(devel)"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "1319dfa5d32891c83be0da597dbe0a8385bb76cb"},
			},
		}, true
	}

	if got := CurrentVersion(); got != "dev+1319dfa" {
		t.Fatalf("expected revision fallback, got %q", got)
	}
}
