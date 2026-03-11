"""AgentFlow installer — deploy rules + modules to CLI target directories."""

from __future__ import annotations

import os
import shutil
import sys
import tempfile
from datetime import datetime
from pathlib import Path

from ._constants import (
    AGENTFLOW_MARKER,
    CLI_DISPLAY_NAMES,
    CLI_HOOKS_FILES,
    CLI_SUBAGENT_FILES,
    CLI_TARGETS,
    DEFAULT_PROFILE,
    HOOKS_SUMMARIES,
    PLUGIN_DIR_NAME,
    PROFILES,
    VALID_PROFILES,
    backup_user_file,
    detect_installed_clis,
    get_agentflow_module_path,
    get_agents_md_path,
    get_skill_md_path,
    is_agentflow_file,
    msg,
)

# ── Windows file-locking safe operations ──────────────────────────────────────


def _safe_remove(path: Path) -> bool:
    """Remove a file or directory, handling Windows file-locking.

    If a ``PermissionError`` is raised (common on Windows when a file is locked),
    the path is renamed aside with a timestamp suffix instead of being deleted.
    """
    if not path.exists():
        return True
    try:
        if path.is_dir():
            shutil.rmtree(path)
        else:
            path.unlink()
        return True
    except PermissionError:
        # Rename-aside fallback for locked files (Windows)
        ts = datetime.now().strftime("%Y%m%d%H%M%S")
        aside = path.parent / f"{path.name}._agentflow_old_{ts}"
        try:
            path.rename(aside)
            print(msg(f"  ⚠️ 文件被锁定，已重命名: {aside.name}", f"  ⚠️ File locked, renamed aside: {aside.name}"))
            return True
        except Exception:
            print(msg(f"  ⚠️ 无法移除: {path}", f"  ⚠️ Cannot remove: {path}"))
            return False


def _safe_write(file_path: Path, content: str) -> bool:
    """Write content to a file atomically via temp file + os.replace.

    This prevents partial writes and handles Windows locking issues.
    """
    try:
        fd, tmp_path = tempfile.mkstemp(
            dir=str(file_path.parent),
            suffix=".tmp",
            prefix=".agentflow_",
        )
        try:
            with os.fdopen(fd, "w", encoding="utf-8") as f:
                f.write(content)
            os.replace(tmp_path, str(file_path))
            return True
        except Exception:
            # Clean up temp file on failure
            try:
                os.unlink(tmp_path)
            except OSError:
                pass
            raise
    except PermissionError:
        # Fallback: direct write if os.replace fails (Windows locking)
        try:
            file_path.write_text(content, encoding="utf-8")
            return True
        except Exception as e:
            print(msg(f"  ⚠️ 写入失败: {e}", f"  ⚠️ Write failed: {e}"))
            return False
    except Exception as e:
        print(msg(f"  ⚠️ 写入失败: {e}", f"  ⚠️ Write failed: {e}"))
        return False


def _get_source_files() -> dict[str, Path]:
    """Get all source files to deploy."""
    module_path = get_agentflow_module_path()
    sources: dict[str, Path] = {}

    for subdir_name in ("stages", "services", "rules", "rlm", "functions", "templates", "hooks", "agents", "core"):
        subdir = module_path / subdir_name
        if subdir.exists():
            sources[subdir_name] = subdir

    return sources


def _build_core_module_for_target(mod_file: str, target: str) -> str:
    """Build a core module's content with CLI-specific sections for the target.

    For ``subagent.md`` and ``hooks.md``, the generic template contains
    placeholders that are replaced with CLI-specific content.
    Other module files are returned as-is.
    """
    core_dir = get_agentflow_module_path() / "core"
    mod_path = core_dir / mod_file
    if not mod_path.exists():
        return ""

    content = mod_path.read_text(encoding="utf-8")

    if mod_file == "subagent.md":
        # Replace {CLI_SUBAGENT_PROTOCOL} with CLI-specific subagent protocol
        cli_file = CLI_SUBAGENT_FILES.get(target, "subagent_other.md")
        cli_path = core_dir / cli_file
        if cli_path.exists():
            cli_content = cli_path.read_text(encoding="utf-8")
        else:
            cli_content = ""
        content = content.replace("{CLI_SUBAGENT_PROTOCOL}", cli_content)

    elif mod_file == "hooks.md":
        # Replace {HOOKS_MATRIX} with CLI-specific hooks matrix
        cli_file = CLI_HOOKS_FILES.get(target, "hooks_other.md")
        cli_path = core_dir / cli_file
        if cli_path.exists():
            cli_content = cli_path.read_text(encoding="utf-8")
        else:
            cli_content = ""
        content = content.replace("{HOOKS_MATRIX}", cli_content)

    return content


def _build_agents_md_for_profile(profile: str, target: str = "codex") -> str:
    """Read the base AGENTS.md and append core extension modules per profile.

    The base AGENTS.md contains G1-G5 (core rules).  The profile determines
    which G6-G12 extension modules from ``agentflow/core/`` are appended.

    Template placeholders are replaced with target-specific content:
    - ``{TARGET_CLI}`` → CLI display name
    - ``{HOOKS_SUMMARY}`` → CLI-specific hooks summary
    """
    source_agents_md = get_agents_md_path()
    content = source_agents_md.read_text(encoding="utf-8")

    marker_line = f"<!-- {AGENTFLOW_MARKER} v1.0.0 -->\n"
    if AGENTFLOW_MARKER not in content:
        content = marker_line + content

    # Replace template placeholders with target-specific values
    display_name = CLI_DISPLAY_NAMES.get(target, target)
    content = content.replace("{TARGET_CLI}", display_name)
    hooks_summary = HOOKS_SUMMARIES.get(target, "Hooks 不可用时功能降级但不影响核心工作流。")
    content = content.replace("{HOOKS_SUMMARY}", hooks_summary)

    # Append core extension modules for this profile
    core_dir = get_agentflow_module_path() / "core"
    modules = PROFILES.get(profile, PROFILES[DEFAULT_PROFILE])

    if modules and core_dir.exists():
        content += "\n\n---\n\n"
        content += f"<!-- PROFILE:{profile} — Extended modules appended below -->\n\n"
        for mod_file in modules:
            mod_content = _build_core_module_for_target(mod_file, target)
            if mod_content:
                content += mod_content + "\n\n"

    return content


def _deploy_rules_file(target: str, cli_dir: Path, profile: str = DEFAULT_PROFILE) -> bool:
    """Deploy the rules file (AGENTS.md / CLAUDE.md / etc.) to the CLI config directory."""
    config = CLI_TARGETS[target]
    rules_file = cli_dir / config["rules_file"]
    source_agents_md = get_agents_md_path()

    if not source_agents_md.exists():
        print(
            msg(
                f"  ⚠️ 找不到 AGENTS.md 源文件: {source_agents_md}",
                f"  ⚠️ AGENTS.md source not found: {source_agents_md}",
            )
        )
        return False

    content = _build_agents_md_for_profile(profile, target)

    if rules_file.exists() and not is_agentflow_file(rules_file):
        backup = backup_user_file(rules_file)
        print(msg(f"  📦 已备份原文件: {backup.name}", f"  📦 Backed up existing file: {backup.name}"))

    _safe_write(rules_file, content)
    profile_label = f" [profile={profile}]" if profile != DEFAULT_PROFILE else ""
    print(
        msg(
            f"  ✅ {config['rules_file']} 已部署{profile_label}", f"  ✅ {config['rules_file']} deployed{profile_label}"
        )
    )
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
                _safe_remove(dest)
            shutil.copytree(source_path, dest)
            deployed += 1
        elif source_path.is_file():
            shutil.copy2(source_path, dest)
            deployed += 1

    print(msg(f"  ✅ 模块目录已部署 ({deployed} 个子模块)", f"  ✅ Module directory deployed ({deployed} submodules)"))
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


def _is_agentflow_hook(hook_handler: dict) -> bool:
    """Check if a hook handler was deployed by AgentFlow."""
    desc = hook_handler.get("description", "")
    return desc.startswith(PLUGIN_DIR_NAME) or desc.startswith("agentflow")


def _deploy_hooks(target: str, cli_dir: Path) -> bool:
    """Deploy hooks configuration for supported CLIs.

    Claude Code uses the new record-based hooks format where hooks is an
    object keyed by event type (e.g. PreToolUse, PostToolUse), each
    containing an array of matcher groups with nested hooks arrays.
    """
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
            new_hooks: dict = hooks_config.get("hooks", {})

            # Migrate: if existing hooks is old array format, discard it
            existing_hooks = existing.get("hooks", {})
            if isinstance(existing_hooks, list):
                existing_hooks = {}

            # Merge per event type
            deployed_count = 0
            for event_type, new_groups in new_hooks.items():
                existing_groups: list = existing_hooks.get(event_type, [])
                # Remove old AgentFlow matcher groups
                cleaned = []
                for group in existing_groups:
                    group_hooks = group.get("hooks", [])
                    filtered = [h for h in group_hooks if not _is_agentflow_hook(h)]
                    if filtered:
                        group["hooks"] = filtered
                        cleaned.append(group)
                cleaned.extend(new_groups)
                existing_hooks[event_type] = cleaned
                deployed_count += len(new_groups)

            existing["hooks"] = existing_hooks

            settings_file.write_text(
                json.dumps(existing, indent=2, ensure_ascii=False),
                encoding="utf-8",
            )
            print(msg(f"  ✅ Hooks 已部署 ({deployed_count} 个)", f"  ✅ Hooks deployed ({deployed_count})"))

    elif target == "codex":
        hooks_src = hooks_dir / "codex_hooks.toml"
        if hooks_src.exists():
            config_file = cli_dir / "config.toml"
            if not config_file.exists():
                shutil.copy2(hooks_src, config_file)
            print(msg("  ✅ Hooks 已部署", "  ✅ Hooks deployed"))

    return True


def _deploy_agent_toml_files(cli_dir: Path) -> None:
    """Deploy reviewer.toml and architect.toml to the Codex agents directory.

    Each deployed file is stamped with the AgentFlow marker so that
    ``is_agentflow_file()`` can identify them during uninstallation.
    """
    agents_src = get_agentflow_module_path() / "agents"
    agents_dest = cli_dir / "agents"
    agents_dest.mkdir(parents=True, exist_ok=True)

    for role_file in ("reviewer.toml", "architect.toml"):
        src = agents_src / role_file
        if src.exists():
            content = src.read_text(encoding="utf-8")
            # Stamp with marker so uninstall can verify ownership
            if AGENTFLOW_MARKER not in content:
                content += f"\n# {AGENTFLOW_MARKER} managed agent role\n"
            _safe_write(agents_dest / role_file, content)

    print(msg("  ✅ 子代理角色已部署 (reviewer, architect)", "  ✅ Agent roles deployed (reviewer, architect)"))


def _merge_codex_config(config_file: Path) -> None:
    """Merge [agents] role definitions and [features] multi_agent into config.toml."""
    try:
        import tomllib
    except ModuleNotFoundError:
        import tomli as tomllib  # type: ignore[no-redef]

    config_text = ""
    if config_file.exists():
        config_text = config_file.read_text(encoding="utf-8")

    try:
        existing = tomllib.loads(config_text)
    except Exception:
        existing = {}

    additions: list[str] = []

    # Enable multi_agent feature
    features = existing.get("features", {})
    if not features.get("multi_agent", False):
        if "features" not in existing:
            additions.append("\n[features]\nmulti_agent = true")
        else:
            config_text = config_text.replace("[features]", "[features]\nmulti_agent = true")

    # Add agent role definitions
    agents = existing.get("agents", {})
    if "reviewer" not in agents:
        additions.append(
            "\n[agents.reviewer]\n"
            'description = "AgentFlow code reviewer: security, correctness, test quality analysis."\n'
            'config_file = "agents/reviewer.toml"'
        )
    if "architect" not in agents:
        additions.append(
            "\n[agents.architect]\n"
            'description = "AgentFlow architect: architectural evaluation, dependency analysis."\n'
            'config_file = "agents/architect.toml"'
        )

    if additions:
        config_text += "\n" + "\n".join(additions) + "\n"
        _safe_write(config_file, config_text)

    print(msg("  ✅ 多代理配置已写入 config.toml", "  ✅ Multi-agent config written to config.toml"))


def _prompt_multi_agent(config_file: Path) -> bool:
    """Prompt the user to enable multi-agent and check existing state.

    Returns ``True`` if multi-agent should be configured, ``False`` to skip.
    """
    try:
        import tomllib
    except ModuleNotFoundError:
        import tomli as tomllib  # type: ignore[no-redef]

    if config_file.exists():
        try:
            existing = tomllib.loads(config_file.read_text(encoding="utf-8"))
            if existing.get("features", {}).get("multi_agent", False):
                print(msg("  ℹ️  多代理已启用，更新角色配置...", "  ℹ️  Multi-agent already enabled, updating roles..."))
                return True
        except Exception:
            pass

    if not sys.stdin.isatty():
        return False

    print()
    print(
        msg(
            "  ┌─────────────────────────────────────────────────┐",
            "  ┌─────────────────────────────────────────────────┐",
        )
    )
    print(
        msg(
            "  │  🤖 多代理协作 (Multi-Agent, 实验性)            │",
            "  │  🤖 Multi-Agent Collaboration (Experimental)    │",
        )
    )
    print(
        msg(
            "  └─────────────────────────────────────────────────┘",
            "  └─────────────────────────────────────────────────┘",
        )
    )
    print(msg("  开启后的能力:", "  When enabled:"))
    print(
        msg(
            "    • reviewer 子代理 — 自动并行审查代码安全性和质量",
            "    • reviewer agent — parallel code security & quality review",
        )
    )
    print(
        msg(
            "    • architect 子代理 — 评估架构方案，对比多种设计",
            "    • architect agent — evaluate architecture, compare designs",
        )
    )
    print(
        msg(
            "    • 多个子代理可并行工作，大幅加速复杂任务",
            "    • Multiple agents work in parallel, speeding up complex tasks",
        )
    )
    print(
        msg("  适用场景: 大型重构、多模块开发、代码审查", "  Best for: large refactors, multi-module dev, code review")
    )
    print(
        msg(
            "  注意: 此功能为实验性，可随时通过 /experimental 关闭",
            "  Note: Experimental feature, can be disabled via /experimental",
        )
    )
    print()
    answer = input(msg("  是否启用多代理？(y/N): ", "  Enable multi-agent? (y/N): ")).strip().lower()
    if answer not in ("y", "yes", "是"):
        print(
            msg(
                "  ⏭️  跳过多代理配置（后续可通过 /experimental 手动开启）",
                "  ⏭️  Skipped (enable later via /experimental in Codex)",
            )
        )
        return False
    return True


def _deploy_codex_agents(cli_dir: Path) -> bool:
    """Deploy AgentFlow agent roles to Codex CLI (multi-agent support).

    Prompts the user whether to enable multi-agent, then:
    1. Deploys reviewer.toml and architect.toml to ~/.codex/agents/
    2. Merges [agents] role definitions into ~/.codex/config.toml
    3. Enables [features] multi_agent = true
    """
    config_file = cli_dir / "config.toml"

    if not _prompt_multi_agent(config_file):
        return True

    _deploy_agent_toml_files(cli_dir)
    _merge_codex_config(config_file)
    return True


# ── Public API ────────────────────────────────────────────────────────────────


def install(target: str, profile: str = DEFAULT_PROFILE) -> bool:
    """Install AgentFlow to a single CLI target.

    Args:
        target: CLI target name (codex, claude, gemini, etc.).
        profile: Deployment profile — ``lite``, ``standard``, or ``full``.
    """
    if target not in CLI_TARGETS:
        print(msg(f"  ❌ 未知目标: {target}", f"  ❌ Unknown target: {target}"))
        print(msg(f"  可用目标: {', '.join(CLI_TARGETS)}", f"  Available: {', '.join(CLI_TARGETS)}"))
        return False

    if profile not in VALID_PROFILES:
        print(msg(f"  ❌ 未知 Profile: {profile}", f"  ❌ Unknown profile: {profile}"))
        print(msg(f"  可用 Profile: {', '.join(VALID_PROFILES)}", f"  Available profiles: {', '.join(VALID_PROFILES)}"))
        return False

    config = CLI_TARGETS[target]
    cli_dir = Path.home() / config["dir"]

    if not cli_dir.exists():
        print(msg(f"  ⚠️ {target} 目录不存在: {cli_dir}", f"  ⚠️ {target} directory not found: {cli_dir}"))
        print(msg(f"  请先安装 {target}。", f"  Please install {target} first."))
        return False

    profile_label = f" (profile={profile})" if profile != DEFAULT_PROFILE else ""
    print(msg(f"\n  正在安装到 {target}{profile_label}...", f"\n  Installing to {target}{profile_label}..."))

    ok = True
    ok = ok and _deploy_rules_file(target, cli_dir, profile)
    ok = ok and _deploy_module_dir(target, cli_dir)
    ok = ok and _deploy_skill_md(target, cli_dir)
    ok = ok and _deploy_hooks(target, cli_dir)

    # Codex-specific: multi-agent setup
    if target == "codex":
        ok = ok and _deploy_codex_agents(cli_dir)

    if ok:
        print(msg(f"\n  ✅ {target} 安装完成!", f"\n  ✅ {target} installation complete!"))
    else:
        print(msg(f"\n  ⚠️ {target} 安装部分失败", f"\n  ⚠️ {target} installation partially failed"))
    return ok


def install_all(profile: str = DEFAULT_PROFILE) -> bool:
    """Install AgentFlow to all detected CLIs."""
    detected = detect_installed_clis()
    if not detected:
        print(msg("  未检测到任何已安装的 CLI。", "  No CLIs detected."))
        return False

    print(
        msg(
            f"  检测到 {len(detected)} 个 CLI: {', '.join(detected)}",
            f"  Detected {len(detected)} CLIs: {', '.join(detected)}",
        )
    )

    success = 0
    for target in detected:
        if install(target, profile):
            success += 1

    print(
        msg(
            f"\n  完成: {success}/{len(detected)} 个目标安装成功",
            f"\n  Done: {success}/{len(detected)} targets installed",
        )
    )
    return success > 0


def _clean_codex_config(config_file: Path) -> None:
    """Remove AgentFlow-injected sections from Codex config.toml.

    Cleans up ``[agents.reviewer]``, ``[agents.architect]`` sections and
    the ``multi_agent = true`` line under ``[features]``.
    """
    if not config_file.exists():
        return

    import re

    text = config_file.read_text(encoding="utf-8")
    original = text

    # Remove [agents.reviewer] and [agents.architect] blocks
    # Each block starts with the header and ends before the next [section] or EOF
    for role in ("reviewer", "architect"):
        pattern = rf"\n?\[agents\.{role}\]\n(?:[^\[]*?)(?=\n\[|\Z)"
        text = re.sub(pattern, "", text)

    # Remove multi_agent = true line (only if under [features])
    text = re.sub(r"\nmulti_agent\s*=\s*true\s*", "\n", text)

    # Remove empty [features] section left behind
    text = re.sub(r"\n\[features\]\s*\n(?=\n|\[|\Z)", "\n", text)

    # Clean up excessive blank lines
    text = re.sub(r"\n{3,}", "\n\n", text).strip() + "\n"

    if text != original:
        _safe_write(config_file, text)
        print(msg("  ✅ config.toml 已清理", "  ✅ config.toml cleaned"))


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
        _safe_remove(rules_file)
        print(msg(f"  ✅ {config['rules_file']} 已移除", f"  ✅ {config['rules_file']} removed"))

    plugin_dir = cli_dir / PLUGIN_DIR_NAME
    if plugin_dir.exists():
        _safe_remove(plugin_dir)
        print(msg("  ✅ 模块目录已移除", "  ✅ Module directory removed"))

    skill_dir = cli_dir / "skills" / "agentflow"
    if skill_dir.exists():
        _safe_remove(skill_dir)
        print(msg("  ✅ SKILL.md 已移除", "  ✅ SKILL.md removed"))

    # Codex-specific: clean up agent role files + config.toml
    if target == "codex":
        agents_dir = cli_dir / "agents"
        removed_any = False
        for role_file in ("reviewer.toml", "architect.toml"):
            rf = agents_dir / role_file
            if rf.exists() and is_agentflow_file(rf):
                _safe_remove(rf)
                removed_any = True
            elif rf.exists():
                print(msg(f"  ⏭️  跳过非 AgentFlow 文件: {role_file}", f"  ⏭️  Skipped non-AgentFlow file: {role_file}"))
        if removed_any:
            print(msg("  ✅ 子代理角色文件已移除", "  ✅ Agent role files removed"))
        # Clean up config.toml entries injected by AgentFlow
        _clean_codex_config(cli_dir / "config.toml")

    if target == "claude":
        settings_file = cli_dir / "settings.json"
        if settings_file.exists():
            import json

            try:
                settings = json.loads(settings_file.read_text(encoding="utf-8"))
                hooks = settings.get("hooks", {})
                if isinstance(hooks, list):
                    # Legacy array format — just remove entirely
                    settings["hooks"] = {}
                elif isinstance(hooks, dict):
                    # New record format — filter AgentFlow entries per event type
                    cleaned_hooks = {}
                    for event_type, groups in hooks.items():
                        cleaned_groups = []
                        for group in groups:
                            group_hooks = group.get("hooks", [])
                            filtered = [h for h in group_hooks if not _is_agentflow_hook(h)]
                            if filtered:
                                group["hooks"] = filtered
                                cleaned_groups.append(group)
                        if cleaned_groups:
                            cleaned_hooks[event_type] = cleaned_groups
                    settings["hooks"] = cleaned_hooks
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
