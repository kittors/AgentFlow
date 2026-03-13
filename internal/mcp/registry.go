package mcp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type BuiltinSpec struct {
	Name        string
	Description string
	Pinned      bool
	Config      map[string]any
}

func BuiltinServers() []BuiltinSpec {
	return []BuiltinSpec{
		{
			Name:        "context7",
			Description: "Up-to-date library/API docs via Context7 (recommended).",
			Pinned:      true,
			Config: map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@upstash/context7-mcp@latest"},
			},
		},
		{
			Name:        "playwright",
			Description: "Browser automation via Playwright MCP (recommended).",
			Pinned:      true,
			Config: map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@playwright/mcp@latest"},
			},
		},
		{
			Name:        "filesystem",
			Description: "Secure file operations with allowlist paths.",
			Pinned:      true,
			Config: map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@modelcontextprotocol/server-filesystem"},
			},
		},
	}
}

func ResolveBuiltin(name string, options InstallOptions) (BuiltinSpec, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return BuiltinSpec{}, errors.New("missing server name")
	}

	for _, spec := range BuiltinServers() {
		if spec.Name != name {
			continue
		}

		switch spec.Name {
		case "context7":
			env := parseEnv(options.Env)
			if len(env) > 0 {
				spec.Config["env"] = env
			}
		case "filesystem":
			allow := make([]any, 0, len(options.Allow))
			for _, entry := range options.Allow {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				allow = append(allow, filepath.Clean(entry))
			}
			if len(allow) > 0 {
				spec.Config["args"] = append(spec.Config["args"].([]any), allow...)
			} else {
				return BuiltinSpec{}, errors.New("filesystem requires --allow=<path> (at least one)")
			}
		default:
		}

		return spec, nil
	}
	return BuiltinSpec{}, fmt.Errorf("unknown server: %s", name)
}

func parseEnv(values []string) map[string]string {
	env := map[string]string{}
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		key, value, ok := strings.Cut(raw, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		env[key] = value
	}
	return env
}
