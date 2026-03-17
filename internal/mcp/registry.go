package mcp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		case "tavily-custom":
			env := parseEnv(options.Env)
			if _, ok := env["TAVILY_API_URL"]; !ok {
				return BuiltinSpec{}, errors.New("tavily-custom requires --set-env=TAVILY_API_URL=<url>")
			}
			if _, ok := env["TAVILY_API_KEY"]; !ok {
				return BuiltinSpec{}, errors.New("tavily-custom requires --set-env=TAVILY_API_KEY=<key>")
			}
			// Ensure scripts are deployed to ~/.agentflow/scripts/tavily-custom-mcp/
			if err := EnsureTavilyCustomScripts(); err != nil {
				return BuiltinSpec{}, fmt.Errorf("failed to deploy tavily-custom scripts: %w", err)
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

// ScriptDir returns the path to the AgentFlow global scripts directory.
// Uses ~/.agentflow/scripts/ as the canonical location.
func ScriptDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".agentflow", "scripts")
	}
	return filepath.Join(".", "scripts")
}

// EnsureTavilyCustomScripts deploys the tavily-custom MCP server scripts
// to ~/.agentflow/scripts/tavily-custom-mcp/ and runs npm install if needed.
func EnsureTavilyCustomScripts() error {
	dir := filepath.Join(ScriptDir(), "tavily-custom-mcp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	// Write index.js
	indexPath := filepath.Join(dir, "index.js")
	if err := os.WriteFile(indexPath, []byte(tavilyCustomIndexJS), 0o644); err != nil {
		return fmt.Errorf("write index.js: %w", err)
	}

	// Write package.json
	pkgPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(tavilyCustomPackageJSON), 0o644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}

	// Run npm install if node_modules is missing
	nmDir := filepath.Join(dir, "node_modules")
	if _, err := os.Stat(nmDir); os.IsNotExist(err) {
		cmd := exec.Command("npm", "install", "--production", "--no-audit", "--no-fund")
		cmd.Dir = dir
		cmd.Stdout = os.Stderr // route npm output to stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install: %w", err)
		}
	}

	return nil
}

// Embedded tavily-custom-mcp scripts.
const tavilyCustomPackageJSON = `{
    "name": "tavily-custom-mcp",
    "version": "1.0.0",
    "description": "Custom Tavily MCP server with configurable proxy URL",
    "type": "module",
    "main": "index.js",
    "bin": {
        "tavily-custom-mcp": "index.js"
    },
    "dependencies": {
        "@modelcontextprotocol/sdk": "^1.12.1",
        "zod": "^3.24.0"
    }
}
`

const tavilyCustomIndexJS = `#!/usr/bin/env node
// tavily-custom-mcp — MCP server for custom Tavily proxy.
// Uses @modelcontextprotocol/sdk for proper protocol handling.

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import http from "http";
import https from "https";

const TAVILY_API_URL = (process.env.TAVILY_API_URL || "").replace(/\/+$/, "");
const TAVILY_API_KEY = process.env.TAVILY_API_KEY || "";

if (!TAVILY_API_URL) {
    console.error("ERROR: TAVILY_API_URL environment variable is required.");
    process.exit(1);
}
if (!TAVILY_API_KEY) {
    console.error("ERROR: TAVILY_API_KEY environment variable is required.");
    process.exit(1);
}

// ---------------------------------------------------------------------------
// HTTP client
// ---------------------------------------------------------------------------

function httpRequest(url, body) {
    return new Promise((resolve, reject) => {
        const parsed = new URL(url);
        const mod = parsed.protocol === "https:" ? https : http;
        const req = mod.request(
            {
                hostname: parsed.hostname,
                port: parsed.port || (parsed.protocol === "https:" ? 443 : 80),
                path: parsed.pathname + parsed.search,
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": ` + "`" + `Bearer ${TAVILY_API_KEY}` + "`" + `,
                },
            },
            (res) => {
                const chunks = [];
                res.on("data", (c) => chunks.push(c));
                res.on("end", () => {
                    const text = Buffer.concat(chunks).toString("utf-8");
                    if (res.statusCode >= 400) reject(new Error(` + "`" + `HTTP ${res.statusCode}: ${text}` + "`" + `));
                    else resolve(text);
                });
            }
        );
        req.on("error", reject);
        req.write(JSON.stringify(body));
        req.end();
    });
}

// ---------------------------------------------------------------------------
// MCP Server
// ---------------------------------------------------------------------------

const server = new McpServer({
    name: "tavily-custom",
    version: "1.0.0",
});

server.tool(
    "tavily-search",
    "Search the web using Tavily AI search. Returns relevant results with titles, URLs, and content snippets.",
    {
        query: z.string().describe("The search query."),
        search_depth: z.enum(["basic", "advanced"]).optional().describe("Search depth: basic or advanced."),
        max_results: z.number().optional().describe("Maximum number of results. Default: 5."),
        include_answer: z.boolean().optional().describe("Include AI-generated answer summary."),
    },
    async (args) => {
        const body = { query: args.query, api_key: TAVILY_API_KEY };
        if (args.search_depth) body.search_depth = args.search_depth;
        if (args.max_results) body.max_results = args.max_results;
        if (args.include_answer !== undefined) body.include_answer = args.include_answer;

        try {
            const raw = await httpRequest(` + "`" + `${TAVILY_API_URL}/api/search` + "`" + `, body);
            return { content: [{ type: "text", text: raw }] };
        } catch (err) {
            return { content: [{ type: "text", text: ` + "`" + `Error: ${err.message}` + "`" + ` }], isError: true };
        }
    }
);

server.tool(
    "tavily-extract",
    "Extract content from one or more URLs using Tavily.",
    {
        urls: z.array(z.string()).describe("URLs to extract content from."),
    },
    async (args) => {
        const body = { urls: args.urls, api_key: TAVILY_API_KEY };

        try {
            const raw = await httpRequest(` + "`" + `${TAVILY_API_URL}/api/extract` + "`" + `, body);
            return { content: [{ type: "text", text: raw }] };
        } catch (err) {
            return { content: [{ type: "text", text: ` + "`" + `Error: ${err.message}` + "`" + ` }], isError: true };
        }
    }
);

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

async function main() {
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error(` + "`" + `tavily-custom MCP ready (${TAVILY_API_URL})` + "`" + `);
}

main().catch((err) => {
    console.error("Fatal:", err);
    process.exit(1);
});
`

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
