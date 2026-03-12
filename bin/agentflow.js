#!/usr/bin/env node

const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");
const { spawnSync } = require("child_process");

const API_URL = "https://api.github.com/repos/kittors/AgentFlow/releases/latest";

function assetName() {
    const platform = process.platform === "win32" ? "windows" : process.platform;
    const arch = process.arch === "x64" ? "amd64" : process.arch === "arm64" ? "arm64" : null;
    if (!arch || !["linux", "darwin", "windows"].includes(platform)) {
        throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
    }
    return `agentflow-${platform}-${arch}${platform === "windows" ? ".exe" : ""}`;
}

function requestJson(url) {
    return new Promise((resolve, reject) => {
        const req = https.get(url, {
            headers: { "User-Agent": "agentflow-npx" },
        }, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                resolve(requestJson(res.headers.location));
                return;
            }
            if (res.statusCode >= 400) {
                reject(new Error(`Request failed with status ${res.statusCode}`));
                return;
            }
            const chunks = [];
            res.on("data", (chunk) => chunks.push(chunk));
            res.on("end", () => {
                try {
                    resolve(JSON.parse(Buffer.concat(chunks).toString("utf8")));
                } catch (error) {
                    reject(error);
                }
            });
        });
        req.on("error", reject);
    });
}

function downloadFile(url, destination) {
    return new Promise((resolve, reject) => {
        const file = fs.createWriteStream(destination);
        https.get(url, { headers: { "User-Agent": "agentflow-npx" } }, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                file.close();
                fs.unlinkSync(destination);
                resolve(downloadFile(res.headers.location, destination));
                return;
            }
            if (res.statusCode >= 400) {
                reject(new Error(`Download failed with status ${res.statusCode}`));
                return;
            }
            res.pipe(file);
            file.on("finish", () => {
                file.close(() => resolve(destination));
            });
        }).on("error", (error) => {
            try { fs.unlinkSync(destination); } catch {}
            reject(error);
        });
    });
}

async function ensureBinary() {
    const home = os.homedir();
    const cacheDir = path.join(home, ".cache", "agentflow", "npx");
    const releaseInfo = await requestJson(API_URL);
    const versionDir = path.join(cacheDir, releaseInfo.tag_name || "latest");
    const binaryPath = path.join(versionDir, assetName());
    if (fs.existsSync(binaryPath)) {
        return binaryPath;
    }

    fs.mkdirSync(versionDir, { recursive: true });
    const desiredAsset = releaseInfo.assets.find((asset) => asset.name === assetName());
    if (!desiredAsset) {
        throw new Error(`Unable to find release asset ${assetName()}`);
    }
    await downloadFile(desiredAsset.browser_download_url, binaryPath);
    if (process.platform !== "win32") {
        fs.chmodSync(binaryPath, 0o755);
    }
    return binaryPath;
}

async function main() {
    try {
        const binaryPath = await ensureBinary();
        const result = spawnSync(binaryPath, process.argv.slice(2), {
            stdio: "inherit",
            env: process.env,
        });
        process.exit(result.status || 0);
    } catch (error) {
        console.error(`AgentFlow npx bootstrap failed: ${error.message}`);
        process.exit(1);
    }
}

main();
