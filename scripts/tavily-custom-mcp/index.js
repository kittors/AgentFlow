#!/usr/bin/env node
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
                    "Authorization": `Bearer ${TAVILY_API_KEY}`,
                },
            },
            (res) => {
                const chunks = [];
                res.on("data", (c) => chunks.push(c));
                res.on("end", () => {
                    const text = Buffer.concat(chunks).toString("utf-8");
                    if (res.statusCode >= 400) reject(new Error(`HTTP ${res.statusCode}: ${text}`));
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
            const raw = await httpRequest(`${TAVILY_API_URL}/api/search`, body);
            return { content: [{ type: "text", text: raw }] };
        } catch (err) {
            return { content: [{ type: "text", text: `Error: ${err.message}` }], isError: true };
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
            const raw = await httpRequest(`${TAVILY_API_URL}/api/extract`, body);
            return { content: [{ type: "text", text: raw }] };
        } catch (err) {
            return { content: [{ type: "text", text: `Error: ${err.message}` }], isError: true };
        }
    }
);

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------

async function main() {
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error(`tavily-custom MCP ready (${TAVILY_API_URL})`);
}

main().catch((err) => {
    console.error("Fatal:", err);
    process.exit(1);
});
