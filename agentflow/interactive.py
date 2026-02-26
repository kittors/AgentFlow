"""AgentFlow interactive menus — multi-select install/uninstall UI."""

from __future__ import annotations

from ._constants import (
    CLI_TARGETS,
    DEFAULT_PROFILE,
    detect_installed_clis,
    detect_installed_targets,
    msg,
)


def _select_profile() -> str:
    """Prompt user to select a deployment profile."""
    print()
    print(msg("  选择部署 Profile:", "  Select deployment profile:"))
    print()
    print(
        msg(
            "  [1] lite     — 仅核心规则 (~310 行, 最小 token 消耗)",
            "  [1] lite     — Core rules only (~310 lines, minimal tokens)",
        )
    )
    print(
        msg("  [2] standard — + 通用规则/验收/模块加载", "  [2] standard — + common rules, acceptance, module loading")
    )
    print(
        msg(
            "  [3] full     — 全部功能 (含子代理/注意力/Hooks)",
            "  [3] full     — All features (sub-agents, attention, hooks)",
        )
    )
    print()

    profile_map = {"1": "lite", "2": "standard", "3": "full"}

    try:
        choice = input(msg("  请输入编号 (默认 3=full): ", "  Enter number (default 3=full): ")).strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return DEFAULT_PROFILE

    return profile_map.get(choice) or DEFAULT_PROFILE


def interactive_install() -> bool:
    """Show interactive CLI target selection for installation."""
    detected = detect_installed_clis()
    if not detected:
        print(msg("  未检测到任何已安装的 CLI。", "  No CLIs detected."))
        return False

    already = detect_installed_targets()

    # Step 1: Select profile
    profile = _select_profile()

    # Step 2: Select targets
    print()
    print(msg("  检测到以下 CLI:", "  Detected CLIs:"))
    print()

    for i, name in enumerate(detected, 1):
        status = msg(" (已安装)", " (installed)") if name in already else ""
        print(f"  [{i}] {name}{status}")

    print()
    print(msg("  [A] 全部安装", "  [A] Install all"))
    print(msg("  [0] 取消", "  [0] Cancel"))
    print()

    try:
        choice = input(msg("  请输入编号 (可多选, 用逗号分隔): ", "  Enter numbers (comma-separated): ")).strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return False

    if not choice or choice == "0":
        return False

    from .installer import install, install_all

    if choice.upper() == "A":
        return install_all(profile)

    targets: list[str] = []
    for part in choice.split(","):
        part = part.strip()
        try:
            idx = int(part)
            if 1 <= idx <= len(detected):
                targets.append(detected[idx - 1])
        except ValueError:
            if part in CLI_TARGETS:
                targets.append(part)

    if not targets:
        print(msg("  无有效选择。", "  No valid selection."))
        return False

    success = 0
    for t in targets:
        if install(t, profile):
            success += 1

    return success > 0


def interactive_uninstall() -> bool:
    """Show interactive CLI target selection for uninstallation."""
    installed = detect_installed_targets()
    if not installed:
        print(msg("  未检测到已安装的 AgentFlow。", "  No AgentFlow installations found."))
        return False

    print()
    print(msg("  已安装 AgentFlow 的 CLI:", "  CLIs with AgentFlow installed:"))
    print()

    for i, name in enumerate(installed, 1):
        print(f"  [{i}] {name}")

    print()
    print(msg("  [A] 全部卸载", "  [A] Uninstall all"))
    print(msg("  [0] 取消", "  [0] Cancel"))
    print()

    try:
        choice = input(msg("  请输入编号 (可多选, 用逗号分隔): ", "  Enter numbers (comma-separated): ")).strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return False

    if not choice or choice == "0":
        return False

    from .installer import uninstall, uninstall_all

    if choice.upper() == "A":
        return uninstall_all()

    targets: list[str] = []
    for part in choice.split(","):
        part = part.strip()
        try:
            idx = int(part)
            if 1 <= idx <= len(installed):
                targets.append(installed[idx - 1])
        except ValueError:
            if part in CLI_TARGETS:
                targets.append(part)

    if not targets:
        print(msg("  无有效选择。", "  No valid selection."))
        return False

    success = 0
    for t in targets:
        if uninstall(t):
            success += 1

    return success > 0
