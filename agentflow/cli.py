"""AgentFlow CLI — entry point for install / uninstall / update / status.

All shared constants and helpers live in ``_constants.py``.
This file only contains the interactive menu and the ``main()`` dispatcher.
"""

from __future__ import annotations

import sys
from importlib.metadata import version as get_version

from ._constants import CLI_TARGETS, msg

# ── Interactive main menu ─────────────────────────────────────────────────────

def _divider(width: int = 40) -> None:
    print("─" * width)


def _interactive_main() -> None:
    """Show main interactive menu for all operations (loops until exit)."""
    from .interactive import interactive_install, interactive_uninstall
    from .updater import clean, status, update

    _divider()
    try:
        ver = get_version("agentflow")
        print(f"  AgentFlow v{ver}")
    except Exception:
        print("  AgentFlow")
    _divider()

    actions = [
        (msg("安装到 CLI 工具", "Install to CLI targets"), "install"),
        (msg("卸载已安装的 CLI 工具", "Uninstall from CLI targets"), "uninstall"),
        (msg("更新 AgentFlow 包", "Update AgentFlow package"), "update"),
        None,  # separator
        (msg("查看安装状态", "Show installation status"), "status"),
        (msg("清理缓存", "Clean caches"), "clean"),
    ]
    flat_actions: list[tuple[str, str]] = [a for a in actions if a is not None]

    while True:
        print()
        print(msg("  请选择操作:", "  Select an action:"))
        print()
        num = 1
        for item in actions:
            if item is None:
                print("  " + "─" * 30)
                continue
            label, _ = item
            print(f"  [{num}] {label}")
            num += 1
        print()
        print(msg("  [0] 退出", "  [0] Exit"))
        print()

        try:
            choice = input(msg("  请输入编号: ", "  Enter number: ")).strip()
        except (EOFError, KeyboardInterrupt):
            print()
            return

        if not choice or choice == "0":
            return

        try:
            idx = int(choice)
            if idx < 1 or idx > len(flat_actions):
                print(msg("  无效编号。", "  Invalid number."))
                continue
        except ValueError:
            print(msg("  无效输入。", "  Invalid input."))
            continue

        _, action = flat_actions[idx - 1]  # type: ignore[index]

        if action == "install":
            interactive_install()
        elif action == "update":
            update()
            return
        elif action == "uninstall":
            interactive_uninstall()
        elif action == "status":
            status()
        elif action == "clean":
            clean()

        print()
        try:
            pause = input(msg("按 Enter 返回主菜单，输入 0 退出: ",
                              "Press Enter to return, 0 to exit: ")).strip()
            if pause == "0":
                return
        except (EOFError, KeyboardInterrupt):
            print()
            return
        _divider()


# ── Usage ─────────────────────────────────────────────────────────────────────

def print_usage() -> None:
    """Print usage information."""
    print("AgentFlow - Multi-CLI Agent Workflow System")
    print()
    print(msg("用法:", "Usage:"))
    print(msg("  agentflow install <target>    安装到指定 CLI",
              "  agentflow install <target>    Install to a specific CLI"))
    print(msg("  agentflow install --all       安装到所有已检测的 CLI",
              "  agentflow install --all       Install to all detected CLIs"))
    print(msg("  agentflow uninstall <target>  从指定 CLI 卸载",
              "  agentflow uninstall <target>  Uninstall from a specific CLI"))
    print(msg("  agentflow uninstall --all     从所有已安装的 CLI 卸载",
              "  agentflow uninstall --all     Uninstall from all installed CLIs"))
    print(msg("  agentflow update              更新到最新版本",
              "  agentflow update              Update to latest version"))
    print(msg("  agentflow clean               清理已安装目标的缓存",
              "  agentflow clean               Clean caches from installed targets"))
    print(msg("  agentflow status              查看安装状态",
              "  agentflow status              Show installation status"))
    print(msg("  agentflow version             查看版本",
              "  agentflow version             Show version"))
    print()
    print(msg("目标:", "Targets:"))
    for name in CLI_TARGETS:
        print(f"  {name}")
    print()


# ── Main entry point ──────────────────────────────────────────────────────────

def main() -> None:
    """Main entry point."""
    from .installer import install, install_all, uninstall, uninstall_all
    from .interactive import interactive_install, interactive_uninstall
    from .updater import clean, status, update

    for stream in (sys.stdout, sys.stderr):
        if hasattr(stream, "reconfigure"):
            try:
                stream.reconfigure(errors="replace")  # type: ignore[union-attr]
            except Exception:
                pass

    cmd = sys.argv[1] if len(sys.argv) >= 2 else None

    if cmd in ("--help", "-h", "help"):
        print_usage()
        sys.exit(0)

    if cmd == "--check-update":
        from .version_check import check_update
        rest_args = sys.argv[2:] if len(sys.argv) > 2 else []  # type: ignore[index]
        silent = "--silent" in rest_args
        check_update(cache_ttl_hours=24, show_version=not silent)
        sys.exit(0)

    if not cmd:
        _interactive_main()
        sys.exit(0)

    if cmd == "install":
        if len(sys.argv) < 3:
            if not interactive_install():
                sys.exit(1)
        else:
            target = sys.argv[2]
            if target == "--all":
                if not install_all():
                    sys.exit(1)
            else:
                if not install(target):
                    sys.exit(1)
    elif cmd == "uninstall":
        if len(sys.argv) < 3:
            if not interactive_uninstall():
                sys.exit(1)
        else:
            target = sys.argv[2]
            if target == "--all":
                uninstall_all()
            else:
                if not uninstall(target):
                    sys.exit(1)
    elif cmd == "update":
        branch = sys.argv[2] if len(sys.argv) >= 3 else None
        update(branch)
    elif cmd == "clean":
        clean()
    elif cmd == "status":
        status()
    elif cmd == "version":
        from .version_check import check_update
        check_update(show_version=True)
    else:
        print(msg(f"未知命令: {cmd}", f"Unknown command: {cmd}"))
        print_usage()
        sys.exit(1)


if __name__ == "__main__":
    main()
