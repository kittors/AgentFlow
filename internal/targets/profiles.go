package targets

import (
	"fmt"
	"sort"
)

var Profiles = map[string][]string{
	"lite": {},
	"standard": {
		"common.md",
		"module_loading.md",
		"acceptance.md",
	},
	"full": {
		"common.md",
		"module_loading.md",
		"acceptance.md",
		"subagent.md",
		"attention.md",
		"hooks.md",
	},
}

const DefaultProfile = "full"

func ValidProfile(profile string) bool {
	_, ok := Profiles[profile]
	return ok
}

func ValidateProfile(profile string) error {
	if !ValidProfile(profile) {
		return fmt.Errorf("invalid profile: %s", profile)
	}
	return nil
}

func SortedProfiles() []string {
	names := make([]string, 0, len(Profiles))
	for name := range Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
