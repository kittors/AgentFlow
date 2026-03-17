package config

import (
	"path/filepath"
	"strings"
)

func RewriteEmbeddedModuleLinks(text, moduleBase string) string {
	moduleBase = strings.TrimSpace(moduleBase)
	if moduleBase == "" {
		return text
	}

	moduleBase = strings.TrimSuffix(filepath.ToSlash(moduleBase), "/")
	return strings.ReplaceAll(text, "(agentflow/", "("+moduleBase+"/")
}
