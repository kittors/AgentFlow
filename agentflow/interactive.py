"""AgentFlow interactive menus — multi-select install/uninstall UI."""

from __future__ import annotations

from ._constants import CLI_TARGETS, detect_installed_clis, detect_installed_targets, msg


def interactive_install() -> bool:
    """Show interactive CLI target selection for installation."""
    detected = detect_installed_clis()
    if not detected:
        print(msg("  未检测到任何已安装的 CLI。", "  No CLIs detected."))
        return False

    already = detect_installed_targets()

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
        choice = input(msg("  请输入编号 (可多选, 用逗号分隔): ",
                           "  Enter numbers (comma-separated): ")).strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return False

    if not choice or choice == "0":
        return False

    from .installer import install, install_all

    if choice.upper() == "A":
        return install_all()

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
        if install(t):
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
        choice = input(msg("  请输入编号 (可多选, 用逗号分隔): ",
                           "  Enter numbers (comma-separated): ")).strip()
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
