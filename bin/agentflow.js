#!/usr/bin/env node

/**
 * AgentFlow npx bridge — installs the Python package and launches the interactive menu.
 * Usage: npx agentflow [install <target>]
 */

const { execSync, spawnSync } = require("child_process");

const REPO = "https://github.com/kittors/AgentFlow";

function findPython() {
    for (const cmd of ["python3", "python", "py"]) {
        try {
            const ver = execSync(`${cmd} --version 2>&1`, { encoding: "utf-8" }).trim();
            const match = ver.match(/(\d+)\.(\d+)/);
            if (match) {
                const major = parseInt(match[1], 10);
                const minor = parseInt(match[2], 10);
                if (major > 3 || (major === 3 && minor >= 10)) {
                    return { cmd, version: ver };
                }
            }
        } catch { }
    }
    return null;
}

function hasCommand(name) {
    try {
        execSync(`${process.platform === "win32" ? "where" : "which"} ${name}`, {
            stdio: "ignore",
        });
        return true;
    } catch {
        return false;
    }
}

function main() {
    console.log("");
    console.log("  ╔═══════════════════════════════════════╗");
    console.log("  ║        AgentFlow — npx installer      ║");
    console.log("  ╚═══════════════════════════════════════╝");
    console.log("");

    // 1. Find Python
    const python = findPython();
    if (!python) {
        console.error("  ✗ Python >= 3.10 is required but not found.");
        console.error("    Please install Python 3.10+ first.");
        process.exit(1);
    }
    console.log(`  ✓ Found ${python.version}`);

    // 2. Check if agentflow is already installed
    if (hasCommand("agentflow")) {
        console.log("  ✓ agentflow is already installed");
        console.log("");
        const args = process.argv.slice(2);
        const result = spawnSync("agentflow", args.length ? args : [], {
            stdio: "inherit",
        });
        process.exit(result.status || 0);
    }

    // 3. Install via uv or pip
    const hasUv = hasCommand("uv");
    console.log(`  · Installing via ${hasUv ? "uv" : "pip"}...`);
    console.log("");

    let installResult;
    if (hasUv) {
        installResult = spawnSync(
            "uv",
            ["tool", "install", "--from", `git+${REPO}`, "agentflow"],
            { stdio: "inherit" }
        );
    } else {
        installResult = spawnSync(
            python.cmd,
            ["-m", "pip", "install", "--upgrade", `git+${REPO}.git`],
            { stdio: "inherit" }
        );
    }

    if (installResult.status !== 0) {
        console.error("  ✗ Installation failed.");
        process.exit(1);
    }

    console.log("");
    console.log("  ✓ AgentFlow installed successfully!");
    console.log("");

    // 4. Launch interactive menu or pass args
    const args = process.argv.slice(2);
    const result = spawnSync("agentflow", args.length ? args : [], {
        stdio: "inherit",
    });
    process.exit(result.status || 0);
}

main();
