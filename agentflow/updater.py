"""AgentFlow updater — update, status, and cache management."""

from __future__ import annotations

import subprocess
import sys

from ._constants import (
    CLI_TARGETS,
    PLUGIN_DIR_NAME,
    REPO_URL,
    detect_install_method,
    detect_installed_targets,
    msg,
)


def update(branch: str | None = None) -> None:
    """Update AgentFlow to the latest version."""
    method = detect_install_method()
    print(msg("  正在更新 AgentFlow...", "  Updating AgentFlow..."))

    try:
        if method == "uv":
            cmd = ["uv", "tool", "upgrade", "agentflow"]
            if branch:
                cmd = [
                    "uv", "tool", "install", "--force",
                    "--from", f"git+{REPO_URL}@{branch}",
                    "agentflow",
                ]
        else:
            url = f"git+{REPO_URL}.git"
            if branch:
                url += f"@{branch}"
            cmd = [sys.executable, "-m", "pip", "install", "--upgrade", url]

        result = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")

        if result.returncode == 0:
            print(msg("  ✅ 更新成功!", "  ✅ Update successful!"))
            installed = detect_installed_targets()
            if installed:
                print(msg(f"  正在重新部署到 {len(installed)} 个目标...",
                          f"  Re-deploying to {len(installed)} targets..."))
                from .installer import install

                for target in installed:
                    install(target)
        else:
            err_text = result.stderr[:200]  # type: ignore[index]
            print(msg(f"  ❌ 更新失败: {err_text}",
                      f"  ❌ Update failed: {err_text}"))
    except Exception as e:
        print(msg(f"  ❌ 更新失败: {e}", f"  ❌ Update failed: {e}"))


def status() -> None:
    """Show installation status."""
    from importlib.metadata import version as get_version
    from pathlib import Path

    print()
    try:
        ver = get_version("agentflow")
        print(msg(f"  版本: {ver}", f"  Version: {ver}"))
    except Exception:
        print(msg("  版本: 未知", "  Version: unknown"))

    method = detect_install_method()
    print(msg(f"  安装方式: {method}", f"  Install method: {method}"))
    print()

    installed = detect_installed_targets()

    print(msg("  CLI 状态:", "  CLI status:"))
    for name, config in CLI_TARGETS.items():
        cli_dir = Path.home() / config["dir"]
        if name in installed:
            print(f"    ✅ {name}")
        elif cli_dir.exists():
            print(msg(f"    ⬚  {name} (已安装 CLI, 未安装 AgentFlow)",
                      f"    ⬚  {name} (CLI installed, AgentFlow not installed)"))
        else:
            print(msg(f"    ─  {name} (未检测到 CLI)",
                      f"    ─  {name} (CLI not detected)"))
    print()


def clean() -> None:
    """Clean caches from installed targets."""
    import shutil
    from pathlib import Path

    installed = detect_installed_targets()
    if not installed:
        print(msg("  未检测到已安装的 AgentFlow。", "  No AgentFlow installations found."))
        return

    cleaned = 0
    for target in installed:
        config = CLI_TARGETS[target]
        cli_dir = Path.home() / config["dir"]

        cache_dirs = [
            cli_dir / PLUGIN_DIR_NAME / "__pycache__",
            cli_dir / PLUGIN_DIR_NAME / ".cache",
        ]

        for cache_dir in cache_dirs:
            if cache_dir.exists():
                shutil.rmtree(cache_dir)
                cleaned += 1

    if cleaned:
        print(msg(f"  ✅ 已清理 {cleaned} 个缓存目录", f"  ✅ Cleaned {cleaned} cache directories"))
    else:
        print(msg("  无需清理。", "  Nothing to clean."))
