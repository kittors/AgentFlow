package mcp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		{
			Name:        "tavily",
			Description: "AI-powered web search via Tavily MCP (requires TAVILY_API_KEY).",
			Pinned:      true,
			Config: map[string]any{
				"command": "npx",
				"args":    []any{"-y", "tavily-mcp@latest"},
			},
		},
		{
			Name:        "tavily-custom",
			Description: "Tavily search via custom proxy (requires TAVILY_API_URL + TAVILY_API_KEY).",
			Pinned:      true,
			Config: map[string]any{
				"command": "node",
				"args":    []any{},
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
		case "tavily":
			env := parseEnv(options.Env)
			if _, ok := env["TAVILY_API_KEY"]; !ok {
				// If no API key provided via --set-env, still allow install
				// but inject a placeholder so the user knows to set it.
				if len(env) == 0 {
					env = map[string]string{"TAVILY_API_KEY": "${TAVILY_API_KEY}"}
				}
			}
			spec.Config["env"] = env
		case "tavily-custom":
			env := parseEnv(options.Env)
			if _, ok := env["TAVILY_API_URL"]; !ok {
				return BuiltinSpec{}, errors.New("tavily-custom requires --set-env=TAVILY_API_URL=<url>")
			}
			if _, ok := env["TAVILY_API_KEY"]; !ok {
				return BuiltinSpec{}, errors.New("tavily-custom requires --set-env=TAVILY_API_KEY=<key>")
			}
			scriptPath := filepath.Join(ScriptDir(), "tavily-custom-mcp", "index.js")
			spec.Config["args"] = []any{scriptPath}
			spec.Config["env"] = env
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

// ScriptDir returns the path to the AgentFlow built-in scripts directory.
// It looks alongside the running executable first, then falls back to a
// development-time location relative to the source tree.
func ScriptDir() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "scripts")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// Fallback: locate relative to the source file (development mode).
	if _, src, _, ok := runtime.Caller(0); ok {
		dir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(src))), "scripts")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return "scripts"
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
