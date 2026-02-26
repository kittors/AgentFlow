"""AgentFlow installer — deploy rules + modules to CLI target directories."""

from __future__ import annotations

import shutil
from pathlib import Path

from ._constants import (
    AGENTFLOW_MARKER,
    CLI_TARGETS,
    PLUGIN_DIR_NAME,
    backup_user_file,
    detect_installed_clis,
    get_agents_md_path,
    get_agentflow_module_path,
    get_skill_md_path,
    is_agentflow_file,
    msg,
)


def _get_source_files() -> dict[str, Path]:
    """Get all source files to deploy."""
    module_path = get_agentflow_module_path()
    sources: dict[str, Path] = {}

    for subdir_name in ("stages", "services", "rules", "rlm", "functions", "templates", "hooks"):
        subdir = module_path / subdir_name
        if subdir.exists():
            sources[subdir_name] = subdir

    return sources


def _deploy_rules_file(target: str, cli_dir: Path) -> bool:
    """Deploy the rules file (AGENTS.md / CLAUDE.md / etc.) to the CLI config directory."""
    config = CLI_TARGETS[target]
    rules_file = cli_dir / config["rules_file"]
    source_agents_md = get_agents_md_path()

    if not source_agents_md.exists():
        print(msg(f"  ⚠️ 找不到 AGENTS.md 源文件: {source_agents_md}",
                  f"  ⚠️ AGENTS.md source not found: {source_agents_md}"))
        return False

    content = source_agents_md.read_text(encoding="utf-8")

    marker_line = f"<!-- {AGENTFLOW_MARKER} v1.0.0 -->\n"
    if AGENTFLOW_MARKER not in content:
        content = marker_line + content

    if rules_file.exists() and not is_agentflow_file(rules_file):
        backup = backup_user_file(rules_file)
        print(msg(f"  📦 已备份原文件: {backup.name}",
                  f"  📦 Backed up existing file: {backup.name}"))

    rules_file.write_text(content, encoding="utf-8")
    print(msg(f"  ✅ {config['rules_file']} 已部署",
              f"  ✅ {config['rules_file']} deployed"))
    return True


def _deploy_module_dir(target: str, cli_dir: Path) -> bool:
    """Deploy the agentflow module directory to the CLI config directory."""
    plugin_dir = cli_dir / PLUGIN_DIR_NAME
    sources = _get_source_files()

    plugin_dir.mkdir(parents=True, exist_ok=True)

    deployed = 0
    for name, source_path in sources.items():
        dest = plugin_dir / name
        if source_path.is_dir():
            if dest.exists():
                shutil.rmtree(dest)
            shutil.copytree(source_path, dest)
            deployed += 1
        elif source_path.is_file():
            shutil.copy2(source_path, dest)
            deployed += 1

    print(msg(f"  ✅ 模块目录已部署 ({deployed} 个子模块)",
              f"  ✅ Module directory deployed ({deployed} submodules)"))
    return True


def _deploy_skill_md(target: str, cli_dir: Path) -> bool:
    """Deploy SKILL.md to the CLI skills discovery directory."""
    skill_source = get_skill_md_path()
    if not skill_source.exists():
        return True

    skill_dir = cli_dir / "skills" / "agentflow"
    skill_dir.mkdir(parents=True, exist_ok=True)
    shutil.copy2(skill_source, skill_dir / "SKILL.md")
    print(msg("  ✅ SKILL.md 已部署", "  ✅ SKILL.md deployed"))
    return True


def _deploy_hooks(target: str, cli_dir: Path) -> bool:
    """Deploy hooks configuration for supported CLIs."""
    hooks_dir = get_agentflow_module_path() / "hooks"
    if not hooks_dir.exists():
        return True

    if target == "claude":
        hooks_src = hooks_dir / "claude_hooks.json"
        if hooks_src.exists():
            import json

            settings_file = cli_dir / "settings.json"
            existing: dict = {}
            if settings_file.exists():
                try:
                    existing = json.loads(settings_file.read_text(encoding="utf-8"))
                except (json.JSONDecodeError, OSError):
                    pass

            hooks_config = json.loads(hooks_src.read_text(encoding="utf-8"))

            existing_hooks: list = existing.get("hooks", [])
            new_hooks: list = hooks_config.get("hooks", [])
            existing_hooks = [
                h for h in existing_hooks
                if not h.get("description", "").startswith(PLUGIN_DIR_NAME)
            ]
            existing_hooks.extend(new_hooks)
            existing["hooks"] = existing_hooks

            settings_file.write_text(
                json.dumps(existing, indent=2, ensure_ascii=False),
                encoding="utf-8",
            )
            print(msg(f"  ✅ Hooks 已部署 ({len(new_hooks)} 个)",
                      f"  ✅ Hooks deployed ({len(new_hooks)})"))

    elif target == "codex":
        hooks_src = hooks_dir / "codex_hooks.toml"
        if hooks_src.exists():
            config_file = cli_dir / "config.toml"
            if not config_file.exists():
                shutil.copy2(hooks_src, config_file)
            print(msg("  ✅ Hooks 已部署", "  ✅ Hooks deployed"))

    return True


# ── Public API ────────────────────────────────────────────────────────────────

def install(target: str) -> bool:
    """Install AgentFlow to a single CLI target."""
    if target not in CLI_TARGETS:
        print(msg(f"  ❌ 未知目标: {target}", f"  ❌ Unknown target: {target}"))
        print(msg(f"  可用目标: {', '.join(CLI_TARGETS)}",
                  f"  Available: {', '.join(CLI_TARGETS)}"))
        return False

    config = CLI_TARGETS[target]
    cli_dir = Path.home() / config["dir"]

    if not cli_dir.exists():
        print(msg(f"  ⚠️ {target} 目录不存在: {cli_dir}",
                  f"  ⚠️ {target} directory not found: {cli_dir}"))
        print(msg(f"  请先安装 {target}。", f"  Please install {target} first."))
        return False

    print(msg(f"\n  正在安装到 {target}...", f"\n  Installing to {target}..."))

    ok = True
    ok = ok and _deploy_rules_file(target, cli_dir)
    ok = ok and _deploy_module_dir(target, cli_dir)
    ok = ok and _deploy_skill_md(target, cli_dir)
    ok = ok and _deploy_hooks(target, cli_dir)

    if ok:
        print(msg(f"\n  ✅ {target} 安装完成!", f"\n  ✅ {target} installation complete!"))
    else:
        print(msg(f"\n  ⚠️ {target} 安装部分失败", f"\n  ⚠️ {target} installation partially failed"))
    return ok


def install_all() -> bool:
    """Install AgentFlow to all detected CLIs."""
    detected = detect_installed_clis()
    if not detected:
        print(msg("  未检测到任何已安装的 CLI。", "  No CLIs detected."))
        return False

    print(msg(f"  检测到 {len(detected)} 个 CLI: {', '.join(detected)}",
              f"  Detected {len(detected)} CLIs: {', '.join(detected)}"))

    success = 0
    for target in detected:
        if install(target):
            success += 1

    print(msg(f"\n  完成: {success}/{len(detected)} 个目标安装成功",
              f"\n  Done: {success}/{len(detected)} targets installed"))
    return success > 0


def uninstall(target: str) -> bool:
    """Uninstall AgentFlow from a single CLI target."""
    if target not in CLI_TARGETS:
        print(msg(f"  ❌ 未知目标: {target}", f"  ❌ Unknown target: {target}"))
        return False

    config = CLI_TARGETS[target]
    cli_dir = Path.home() / config["dir"]

    print(msg(f"\n  正在从 {target} 卸载...", f"\n  Uninstalling from {target}..."))

    rules_file = cli_dir / config["rules_file"]
    if rules_file.exists() and is_agentflow_file(rules_file):
        rules_file.unlink()
        print(msg(f"  ✅ {config['rules_file']} 已移除",
                  f"  ✅ {config['rules_file']} removed"))

    plugin_dir = cli_dir / PLUGIN_DIR_NAME
    if plugin_dir.exists():
        shutil.rmtree(plugin_dir)
        print(msg("  ✅ 模块目录已移除", "  ✅ Module directory removed"))

    skill_dir = cli_dir / "skills" / "agentflow"
    if skill_dir.exists():
        shutil.rmtree(skill_dir)
        print(msg("  ✅ SKILL.md 已移除", "  ✅ SKILL.md removed"))

    if target == "claude":
        settings_file = cli_dir / "settings.json"
        if settings_file.exists():
            import json

            try:
                settings = json.loads(settings_file.read_text(encoding="utf-8"))
                hooks = settings.get("hooks", [])
                settings["hooks"] = [
                    h for h in hooks
                    if not h.get("description", "").startswith(PLUGIN_DIR_NAME)
                ]
                settings_file.write_text(
                    json.dumps(settings, indent=2, ensure_ascii=False),
                    encoding="utf-8",
                )
                print(msg("  ✅ Hooks 已清理", "  ✅ Hooks cleaned"))
            except Exception:
                pass

    print(msg(f"  ✅ {target} 卸载完成!", f"  ✅ {target} uninstalled!"))
    return True


def uninstall_all() -> bool:
    """Uninstall AgentFlow from all installed targets."""
    from ._constants import detect_installed_targets

    installed = detect_installed_targets()
    if not installed:
        print(msg("  未检测到已安装的 AgentFlow。", "  No AgentFlow installations found."))
        return False

    for target in installed:
        uninstall(target)
    return True
