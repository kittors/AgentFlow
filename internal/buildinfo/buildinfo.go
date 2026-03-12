package buildinfo

import (
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

var readBuildInfo = debug.ReadBuildInfo

func CurrentVersion() string {
	if version := normalizeVersion(Version); version != "" && !isDevVersion(version) {
		return version
	}

	if info, ok := readBuildInfo(); ok {
		if version := normalizeVersion(info.Main.Version); version != "" && !isDevVersion(version) {
			return version
		}
		if revision := lookupSetting(info, "vcs.revision"); revision != "" {
			return "dev+" + shortCommit(revision)
		}
	}

	if version := normalizeVersion(Version); version != "" {
		return version
	}
	return "dev"
}

func CurrentCommit() string {
	if Commit != "" {
		return shortCommit(Commit)
	}
	if info, ok := readBuildInfo(); ok {
		if revision := lookupSetting(info, "vcs.revision"); revision != "" {
			return shortCommit(revision)
		}
	}
	return ""
}

func lookupSetting(info *debug.BuildInfo, key string) string {
	if info == nil {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	switch version {
	case "", "(devel)", "unknown":
		return ""
	default:
		return version
	}
}

func isDevVersion(version string) bool {
	return version == "dev" || strings.HasPrefix(version, "dev+")
}

func shortCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}
