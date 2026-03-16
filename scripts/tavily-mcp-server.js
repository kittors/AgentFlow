#!/usr/bin/env node
// tavily-mcp-server.js — Lightweight MCP server for custom Tavily proxy.
// Zero dependencies: uses only Node.js built-in modules.
//
// Environment variables:
//   TAVILY_API_URL  — base URL of the Tavily-compatible API (required)
//   TAVILY_API_KEY  — API key for authentication (required)

"use strict";

const http = require("http");
const https = require("https");
const readline = require("readline");

const TAVILY_API_URL = (process.env.TAVILY_API_URL || "").replace(/\/+$/, "");
const TAVILY_API_KEY = process.env.TAVILY_API_KEY || "";

// ---------------------------------------------------------------------------
// JSON-RPC helpers
// ---------------------------------------------------------------------------

let nextId = 1;

function jsonrpcResponse(id, result) {
    return JSON.stringify({ jsonrpc: "2.0", id, result });
}

function jsonrpcError(id, code, message, data) {
    const err = { code, message };
    if (data !== undefined) err.data = data;
    return JSON.stringify({ jsonrpc: "2.0", id, error: err });
}

// ---------------------------------------------------------------------------
// HTTP client
// ---------------------------------------------------------------------------

function httpRequest(url, body, headers) {
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
                    ...headers,
                },
            },
            (res) => {
                const chunks = [];
                res.on("data", (c) => chunks.push(c));
                res.on("end", () => {
                    const text = Buffer.concat(chunks).toString("utf-8");
                    if (res.statusCode >= 400) {
                        reject(new Error(`HTTP ${res.statusCode}: ${text}`));
                    } else {
                        resolve(text);
                    }
                });
            }
        );
        req.on("error", reject);
        req.write(JSON.stringify(body));
        req.end();
    });
}

// ---------------------------------------------------------------------------
// Tavily API wrappers
// ---------------------------------------------------------------------------

async function tavilySearch(query, opts) {
    const body = { query, api_key: TAVILY_API_KEY };
    if (opts.search_depth) body.search_depth = opts.search_depth;
    if (opts.max_results) body.max_results = opts.max_results;
    if (opts.include_answer !== undefined) body.include_answer = opts.include_answer;

    const headers = {};
    if (TAVILY_API_KEY) {
        headers["Authorization"] = `Bearer ${TAVILY_API_KEY}`;
    }

    const raw = await httpRequest(`${TAVILY_API_URL}/api/search`, body, headers);
    return JSON.parse(raw);
}

async function tavilyExtract(urls) {
    const body = { urls, api_key: TAVILY_API_KEY };

    const headers = {};
    if (TAVILY_API_KEY) {
        headers["Authorization"] = `Bearer ${TAVILY_API_KEY}`;
    }

    const raw = await httpRequest(`${TAVILY_API_URL}/api/extract`, body, headers);
    return JSON.parse(raw);
}

// ---------------------------------------------------------------------------
// MCP tool definitions
// ---------------------------------------------------------------------------

const TOOLS = [
    {
        name: "tavily-search",
        description:
            "Search the web using Tavily AI search. Returns relevant results with titles, URLs, and content snippets.",
        inputSchema: {
            type: "object",
            properties: {
                query: { type: "string", description: "The search query." },
                search_depth: {
                    type: "string",
                    enum: ["basic", "advanced"],
                    description: "Search depth: basic (fast) or advanced (thorough). Default: basic.",
                },
                max_results: {
                    type: "number",
                    description: "Maximum number of results to return. Default: 5.",
                },
                include_answer: {
                    type: "boolean",
                    description: "Whether to include an AI-generated answer summary. Default: false.",
                },
            },
            required: ["query"],
        },
    },
    {
        name: "tavily-extract",
        description:
            "Extract content from one or more URLs using Tavily. Returns the main text content of each page.",
        inputSchema: {
            type: "object",
            properties: {
                urls: {
                    type: "array",
                    items: { type: "string" },
                    description: "List of URLs to extract content from.",
                },
            },
            required: ["urls"],
        },
    },
];

// ---------------------------------------------------------------------------
// MCP message handler
// ---------------------------------------------------------------------------

async function handleMessage(msg) {
    const { method, id, params } = msg;

    switch (method) {
        case "initialize":
            return jsonrpcResponse(id, {
                protocolVersion: "2024-11-05",
                capabilities: { tools: { listChanged: false } },
                serverInfo: { name: "tavily-custom", version: "1.0.0" },
            });

        case "notifications/initialized":
            // No response needed for notifications.
            return null;

        case "tools/list":
            return jsonrpcResponse(id, { tools: TOOLS });

        case "tools/call": {
            const toolName = params && params.name;
            const args = (params && params.arguments) || {};

            if (toolName === "tavily-search") {
                if (!args.query) {
                    return jsonrpcError(id, -32602, "Missing required parameter: query");
                }
                try {
                    const result = await tavilySearch(args.query, args);
                    return jsonrpcResponse(id, {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
                    });
                } catch (err) {
                    return jsonrpcResponse(id, {
                        content: [{ type: "text", text: `Error: ${err.message}` }],
                        isError: true,
                    });
                }
            }

            if (toolName === "tavily-extract") {
                if (!args.urls || !Array.isArray(args.urls) || args.urls.length === 0) {
                    return jsonrpcError(id, -32602, "Missing required parameter: urls");
                }
                try {
                    const result = await tavilyExtract(args.urls);
                    return jsonrpcResponse(id, {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
                    });
                } catch (err) {
                    return jsonrpcResponse(id, {
                        content: [{ type: "text", text: `Error: ${err.message}` }],
                        isError: true,
                    });
                }
            }

            return jsonrpcError(id, -32601, `Unknown tool: ${toolName}`);
        }

        case "ping":
            return jsonrpcResponse(id, {});

        default:
            if (id !== undefined && id !== null) {
                return jsonrpcError(id, -32601, `Method not found: ${method}`);
            }
            // Ignore unknown notifications.
            return null;
    }
}

// ---------------------------------------------------------------------------
// stdio transport
// ---------------------------------------------------------------------------

function main() {
    if (!TAVILY_API_URL) {
        console.error("ERROR: TAVILY_API_URL environment variable is required.");
        process.exit(1);
    }
    if (!TAVILY_API_KEY) {
        console.error("ERROR: TAVILY_API_KEY environment variable is required.");
        process.exit(1);
    }

    console.error(`Tavily custom MCP server started (API: ${TAVILY_API_URL})`);

    const rl = readline.createInterface({ input: process.stdin, terminal: false });

    let buffer = "";

    process.stdin.on("data", (chunk) => {
        buffer += chunk.toString();

        // MCP uses Content-Length framed messages over stdio.
        while (true) {
            const headerEnd = buffer.indexOf("\r\n\r\n");
            if (headerEnd < 0) break;

            const headerSection = buffer.substring(0, headerEnd);
            const match = headerSection.match(/Content-Length:\s*(\d+)/i);
            if (!match) {
                // Not a valid header — skip past it.
                buffer = buffer.substring(headerEnd + 4);
                continue;
            }

            const contentLength = parseInt(match[1], 10);
            const bodyStart = headerEnd + 4;

            if (buffer.length < bodyStart + contentLength) {
                // Not enough data yet.
                break;
            }

            const body = buffer.substring(bodyStart, bodyStart + contentLength);
            buffer = buffer.substring(bodyStart + contentLength);

            let msg;
            try {
                msg = JSON.parse(body);
            } catch (e) {
                console.error("Failed to parse JSON-RPC message:", e.message);
                continue;
            }

            handleMessage(msg)
                .then((response) => {
                    if (response !== null) {
                        const bytes = Buffer.byteLength(response, "utf-8");
                        process.stdout.write(`Content-Length: ${bytes}\r\n\r\n${response}`);
                    }
                })
                .catch((err) => {
                    console.error("Handler error:", err);
                    if (msg.id !== undefined && msg.id !== null) {
                        const errResp = jsonrpcError(msg.id, -32603, err.message);
                        const bytes = Buffer.byteLength(errResp, "utf-8");
                        process.stdout.write(`Content-Length: ${bytes}\r\n\r\n${errResp}`);
                    }
                });
        }
    });

    // Close readline to avoid it consuming stdin.
    rl.close();

    process.stdin.on("end", () => {
        process.exit(0);
    });
}

main();
