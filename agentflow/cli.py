"""AgentFlow CLI — entry point for install / uninstall / update / status.

All shared constants and helpers live in ``_constants.py``.
This file only contains the interactive menu and the ``main()`` dispatcher.
"""

from __future__ import annotations

import sys
from importlib.metadata import version as get_version

from ._constants import CLI_TARGETS, DEFAULT_PROFILE, VALID_PROFILES, msg

# ── Rich + InquirerPy TUI helpers ─────────────────────────────────────────────

_HAS_TUI = True
try:
    from InquirerPy import inquirer
    from InquirerPy.base.control import Choice
    from InquirerPy.separator import Separator
    from rich.console import Console
    from rich.panel import Panel
    from rich.text import Text
except ImportError:
    _HAS_TUI = False


def _get_console() -> Console:
    """Return a Rich Console instance."""
    return Console()


def _print_banner(console: Console) -> None:
    """Print a beautiful styled banner using Rich."""
    try:
        ver = get_version("agentflow")
    except Exception:
        ver = "dev"

    title = Text()
    title.append("🚀 ", style="bold")
    title.append("AgentFlow", style="bold cyan")
    title.append(f" v{ver}", style="dim cyan")

    subtitle = Text(msg("多 CLI 代理工作流系统", "Multi-CLI Agent Workflow System"), style="dim")

    content = Text()
    content.append_text(title)
    content.append("\n")
    content.append_text(subtitle)

    panel = Panel(
        content,
        border_style="cyan",
        padding=(1, 2),
    )
    console.print()
    console.print(panel)
    console.print()


# ── Interactive main menu (TUI version) ──────────────────────────────────────


def _interactive_main_tui() -> None:
    """Show main interactive menu using Rich + InquirerPy."""
    from .interactive import interactive_install_tui, interactive_uninstall_tui
    from .updater import clean, status, update

    console = _get_console()
    _print_banner(console)

    actions = {
        "install": (msg("📦  安装到 CLI 工具", "📦  Install to CLI targets"), "install"),
        "uninstall": (msg("🗑️   卸载已安装的 CLI 工具", "🗑️   Uninstall from CLI targets"), "uninstall"),
        "update": (msg("🔄  更新 AgentFlow 包", "🔄  Update AgentFlow package"), "update"),
        "status": (msg("📊  查看安装状态", "📊  Show installation status"), "status"),
        "clean": (msg("🧹  清理缓存", "🧹  Clean caches"), "clean"),
        "exit": (msg("🚪  退出", "🚪  Exit"), "exit"),
    }

    while True:
        try:
            choices = [
                Choice("install", name=actions["install"][0]),
                Choice("uninstall", name=actions["uninstall"][0]),
                Choice("update", name=actions["update"][0]),
                Separator(),
                Choice("status", name=actions["status"][0]),
                Choice("clean", name=actions["clean"][0]),
                Separator(),
                Choice("exit", name=actions["exit"][0]),
            ]

            action = inquirer.select(
                message=msg("请选择操作", "Select an action"),
                choices=choices,
                default="install",
                pointer="❯",
                qmark="",
                amark="✔",
                instruction=msg("(↑↓ 选择, Enter 确认)", "(↑↓ to move, Enter to select)"),
            ).execute()

        except (EOFError, KeyboardInterrupt):
            console.print()
            return

        if action == "exit":
            return
        elif action == "install":
            interactive_install_tui()
        elif action == "uninstall":
            interactive_uninstall_tui()
        elif action == "update":
            update()
            return
        elif action == "status":
            status()
        elif action == "clean":
            clean()

        console.print()
        try:
            resp = input(msg("按 Enter 返回主菜单，输入 0 退出: ", "Press Enter to return, 0 to exit: ")).strip()
            if resp == "0":
                return
        except (EOFError, KeyboardInterrupt):
            console.print()
            return


# ── Interactive main menu (fallback plain version) ───────────────────────────


def _interactive_main_plain() -> None:
    """Show main interactive menu using plain text (fallback)."""
    from .interactive import interactive_install, interactive_uninstall
    from .updater import clean, status, update

    print("─" * 40)
    try:
        ver = get_version("agentflow")
        print(f"  AgentFlow v{ver}")
    except Exception:
        print("  AgentFlow")
    print("─" * 40)

    action_items = [
        (msg("安装到 CLI 工具", "Install to CLI targets"), "install"),
        (msg("卸载已安装的 CLI 工具", "Uninstall from CLI targets"), "uninstall"),
        (msg("更新 AgentFlow 包", "Update AgentFlow package"), "update"),
        None,
        (msg("查看安装状态", "Show installation status"), "status"),
        (msg("清理缓存", "Clean caches"), "clean"),
    ]
    flat_actions: list[tuple[str, str]] = [a for a in action_items if a is not None]

    while True:
        print()
        print(msg("  请选择操作:", "  Select an action:"))
        print()
        num = 1
        for item in action_items:
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

        _, action = flat_actions[idx - 1]

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
            pause = input(msg("按 Enter 返回主菜单，输入 0 退出: ", "Press Enter to return, 0 to exit: ")).strip()
            if pause == "0":
                return
        except (EOFError, KeyboardInterrupt):
            print()
            return
        print("─" * 40)


# ── Usage ─────────────────────────────────────────────────────────────────────


def print_usage() -> None:
    """Print usage information."""
    if _HAS_TUI:
        console = _get_console()
        from rich.table import Table

        console.print(
            f"\n[bold cyan]AgentFlow[/] — {msg('多 CLI 代理工作流系统', 'Multi-CLI Agent Workflow System')}\n"
        )

        table = Table(show_header=True, header_style="bold", border_style="dim")
        table.add_column(msg("命令", "Command"), style="cyan", min_width=35)
        table.add_column(msg("说明", "Description"))

        table.add_row("agentflow install <target>", msg("安装到指定 CLI", "Install to a specific CLI"))
        table.add_row("agentflow install --all", msg("安装到所有已检测的 CLI", "Install to all detected CLIs"))
        table.add_row("agentflow uninstall <target>", msg("从指定 CLI 卸载", "Uninstall from a specific CLI"))
        table.add_row("agentflow uninstall --all", msg("从所有已安装的 CLI 卸载", "Uninstall from all installed CLIs"))
        table.add_row("agentflow update", msg("更新到最新版本", "Update to latest version"))
        table.add_row("agentflow clean", msg("清理已安装目标的缓存", "Clean caches from installed targets"))
        table.add_row("agentflow status", msg("查看安装状态", "Show installation status"))
        table.add_row("agentflow version", msg("查看版本", "Show version"))

        console.print(table)

        profiles_str = ", ".join(VALID_PROFILES)
        console.print(f"\n[bold]{msg('选项', 'Options')}:[/]")
        profile_desc = msg(
            f"选择部署 Profile (默认: {DEFAULT_PROFILE})",
            f"Deployment profile (default: {DEFAULT_PROFILE})",
        )
        console.print(f"  --profile=<{profiles_str}>  {profile_desc}")
        console.print(f"\n[bold]{msg('目标', 'Targets')}:[/]")
        for name in CLI_TARGETS:
            console.print(f"  [cyan]{name}[/]")
        console.print()
    else:
        _print_usage_plain()


def _print_usage_plain() -> None:
    """Plain text usage (fallback)."""
    print("AgentFlow - Multi-CLI Agent Workflow System")
    print()
    print(msg("用法:", "Usage:"))
    print(
        msg(
            "  agentflow install <target>    安装到指定 CLI",
            "  agentflow install <target>    Install to a specific CLI",
        )
    )
    print(
        msg(
            "  agentflow install --all       安装到所有已检测的 CLI",
            "  agentflow install --all       Install to all detected CLIs",
        )
    )
    print(
        msg(
            "  agentflow uninstall <target>  从指定 CLI 卸载",
            "  agentflow uninstall <target>  Uninstall from a specific CLI",
        )
    )
    print(
        msg(
            "  agentflow uninstall --all     从所有已安装的 CLI 卸载",
            "  agentflow uninstall --all     Uninstall from all installed CLIs",
        )
    )
    print(
        msg(
            "  agentflow update              更新到最新版本", "  agentflow update              Update to latest version"
        )
    )
    print(
        msg(
            "  agentflow clean               清理已安装目标的缓存",
            "  agentflow clean               Clean caches from installed targets",
        )
    )
    print(
        msg("  agentflow status              查看安装状态", "  agentflow status              Show installation status")
    )
    print(msg("  agentflow version             查看版本", "  agentflow version             Show version"))
    print()
    print(msg("选项:", "Options:"))
    profiles_str = ", ".join(VALID_PROFILES)
    print(
        msg(
            f"  --profile=<{profiles_str}>  选择部署 Profile (默认: {DEFAULT_PROFILE})",
            f"  --profile=<{profiles_str}>  Deployment profile (default: {DEFAULT_PROFILE})",
        )
    )
    print()
    print(msg("目标:", "Targets:"))
    for name in CLI_TARGETS:
        print(f"  {name}")
    print()


# ── Main entry point ──────────────────────────────────────────────────────────


def main() -> None:
    """Main entry point."""
    from .installer import install, install_all, uninstall, uninstall_all
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
        if _HAS_TUI and sys.stdin.isatty():
            try:
                _interactive_main_tui()
            except OSError:
                # prompt_toolkit fails in piped envs (curl|bash)
                _interactive_main_plain()
        else:
            _interactive_main_plain()
        sys.exit(0)

    if cmd == "install":
        rest_args = sys.argv[2:]
        profile = DEFAULT_PROFILE
        non_profile_args: list[str] = []
        for arg in rest_args:
            if arg.startswith("--profile="):
                profile = arg.split("=", 1)[1]
            else:
                non_profile_args.append(arg)

        if not non_profile_args:
            if _HAS_TUI and sys.stdin.isatty():
                from .interactive import interactive_install_tui

                try:
                    if not interactive_install_tui():
                        sys.exit(1)
                except OSError:
                    from .interactive import interactive_install

                    if not interactive_install():
                        sys.exit(1)
            else:
                from .interactive import interactive_install

                if not interactive_install():
                    sys.exit(1)
        else:
            target = non_profile_args[0]
            if target == "--all":
                if not install_all(profile):
                    sys.exit(1)
            else:
                if not install(target, profile):
                    sys.exit(1)
    elif cmd == "uninstall":
        if len(sys.argv) < 3:
            if _HAS_TUI and sys.stdin.isatty():
                from .interactive import interactive_uninstall_tui

                try:
                    if not interactive_uninstall_tui():
                        sys.exit(1)
                except OSError:
                    from .interactive import interactive_uninstall

                    if not interactive_uninstall():
                        sys.exit(1)
            else:
                from .interactive import interactive_uninstall

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
    else:

        def _cmd_clean() -> None:
            clean()

        def _cmd_status() -> None:
            status()

        def _cmd_version() -> None:
            from .version_check import check_update

            check_update(show_version=True)

        simple_commands = {
            "clean": _cmd_clean,
            "status": _cmd_status,
            "version": _cmd_version,
        }

        handler = simple_commands.get(cmd)  # type: ignore[arg-type]
        if handler:
            handler()
        else:
            print(msg(f"未知命令: {cmd}", f"Unknown command: {cmd}"))
            print_usage()
            sys.exit(1)


if __name__ == "__main__":
    main()
