#!/usr/bin/env bash
set -eu

# ─── Configuration ───
REPO="${AGENTFLOW_REPO:-https://github.com/kittors/AgentFlow}"
BRANCH="${AGENTFLOW_BRANCH:-main}"

# ─── Colors ───
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

# ─── Language selection ───
# Auto-detect first, then allow override via interactive prompt
USE_ZH=false
_locale="${LC_ALL:-${LC_MESSAGES:-${LANG:-${LANGUAGE:-}}}}"
case "$_locale" in
    zh*|ZH*) USE_ZH=true ;;
esac

# Interactive language selection (only if stdin is a tty)
if [ -t 0 ] 2>/dev/null || [ -e /dev/tty ]; then
    printf "\n"
    printf "  ${BOLD}Select language / 选择语言:${RESET}\n"
    printf "  [1] 中文\n"
    printf "  [2] English\n"
    printf "\n"
    _lang_choice=""
    if [ -t 0 ]; then
        printf "  (1/2): " && read -r _lang_choice
    elif [ -e /dev/tty ]; then
        printf "  (1/2): " && read -r _lang_choice </dev/tty
    fi
    case "$_lang_choice" in
        1) USE_ZH=true ;;
        2) USE_ZH=false ;;
        # Empty or other: keep auto-detected value
    esac
fi

msg() {
    if [ "$USE_ZH" = true ]; then echo "$1"; else echo "$2"; fi
}

info()    { printf "${CYAN}  [·]${RESET}  %s\n" "$*"; }
ok()      { printf "${GREEN}  [✓]${RESET}  %s\n" "$*"; }
warn()    { printf "${YELLOW}  [!]${RESET}  %s\n" "$*"; }
err()     { printf "${RED}  [✗]${RESET}  %s\n" "$*"; exit 1; }
step()    { printf "\n${BOLD}${CYAN}  ─── %s ───${RESET}\n\n" "$*"; }

# ─── Ensure PATH in shell profile ───
_ensure_path_in_profile() {
    local bin_dir="$1"
    local line="export PATH=\"${bin_dir}:\$PATH\""
    local updated=false

    # Determine which shell profile(s) to update
    local profiles=""
    local current_shell="${SHELL:-}"
    case "$current_shell" in
        */zsh)  profiles="$HOME/.zshrc" ;;
        */bash) profiles="$HOME/.bashrc $HOME/.bash_profile" ;;
        *)      profiles="$HOME/.profile" ;;
    esac

    for profile in $profiles; do
        # Create profile if it doesn't exist
        [ -f "$profile" ] || touch "$profile"
        # Check if already present
        if ! grep -qF "$bin_dir" "$profile" 2>/dev/null; then
            printf "\n# Added by AgentFlow installer\n%s\n" "$line" >> "$profile"
            updated=true
            ok "$(msg "PATH 已写入 $(basename $profile)" "PATH added to $(basename $profile)")"
        fi
    done

    if [ "$updated" = false ]; then
        ok "$(msg "PATH 已存在于 shell 配置中" "PATH already configured in shell profile")"
    fi
}

# ─── Banner ───
printf "\n"
printf "${BOLD}${CYAN}"
printf "     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗\n"
printf "    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║\n"
printf "    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║\n"
printf "    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║\n"
printf "    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝\n"
printf "    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚═╝╚═╝ \n"
printf "${RESET}"
printf "${DIM}    Multi-CLI Agent Workflow System${RESET}\n"
printf "\n"

# ─── Step 1: Detect Git ───
step "$(msg '步骤 1/5: 检测 Git' 'Step 1/5: Detecting Git')"

if ! command -v git >/dev/null 2>&1; then
    err "$(msg 'Git 未找到。请先安装 Git。' 'Git not found. Please install Git first.')"
fi
ok "$(msg "找到 Git ($(git --version 2>&1))" "Found Git ($(git --version 2>&1))")"

# ─── Step 2: Detect package manager ───
step "$(msg '步骤 2/5: 检测包管理器' 'Step 2/5: Detecting package manager')"

HAS_UV=false
if command -v uv >/dev/null 2>&1; then
    HAS_UV=true
    ok "$(msg "找到 uv ($(uv --version 2>&1))，将优先使用" "Found uv ($(uv --version 2>&1)), will use it")"
else
    info "$(msg '未找到 uv，将使用 pip' 'uv not found, falling back to pip')"
fi

# ─── Step 3: Detect Python ───
step "$(msg '步骤 3/5: 检测 Python' 'Step 3/5: Detecting Python')"

PYTHON_CMD=""
# 检查带具体版本号的 python3 或者 python
for cmd in python3.14 python3.13 python3.12 python3.11 python3.10 python3 python; do
    if command -v "$cmd" >/dev/null 2>&1; then
        # 兼容例如 Python 3.10.x 等格式输出
        version=$("$cmd" --version 2>&1 | grep -oE '[0-9]+\.[0-9]+' | head -1)
        if [ -n "$version" ]; then
            major=$(echo "$version" | cut -d. -f1)
            minor=$(echo "$version" | cut -d. -f2)
            if [ "$major" -gt 3 ] || { [ "$major" -eq 3 ] && [ "$minor" -ge 10 ]; }; then
                PYTHON_CMD="$cmd"
                break
            fi
        fi
    fi
done

if [ -z "$PYTHON_CMD" ]; then
    if [ "$HAS_UV" = true ]; then
        warn "$(msg '未找到 Python >= 3.10，但已安装 uv。将使用 uv 继续安装。' 'Python >= 3.10 not found, but uv is installed. Proceeding with uv.')"
    else
        err "$(msg '需要 Python >= 3.10，但未找到。请安装后重试。' 'Python >= 3.10 is required but not found. Please install it and try again.')"
    fi
else
    ok "$(msg "找到 $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))" "Found $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))")"
fi

# ─── Step 3.5: Clean pip remnants ───
if [ -n "$PYTHON_CMD" ]; then
    while IFS= read -r sp_dir; do
        [ -d "$sp_dir" ] || continue
        for remnant in "$sp_dir"/~*; do
            [ -e "$remnant" ] || continue
            if rm -rf "$remnant" 2>/dev/null; then
                info "$(msg "已清理 pip 残留: $(basename "$remnant")" "Cleaned pip remnant: $(basename "$remnant")")"
            fi
        done
    done < <("$PYTHON_CMD" -c "import site
for p in site.getsitepackages():
    print(p)" 2>/dev/null || true)
fi

# ─── Step 4: Install ───
step "$(msg "步骤 4/5: 安装 AgentFlow (分支: $BRANCH)" "Step 4/5: Installing AgentFlow (branch: $BRANCH)")"

if [ "$HAS_UV" = true ]; then
    info "$(msg '使用 uv 安装...' 'Installing with uv...')"
    if [ "$BRANCH" = "main" ]; then
        uv tool install --force --from "git+${REPO}" agentflow
    else
        uv tool install --force --from "git+${REPO}@${BRANCH}" agentflow
    fi
    # Ensure ~/.local/bin is on PATH for the rest of this script
    export PATH="$HOME/.local/bin:$PATH"
    # Persist PATH into shell config (manual write, uv tool update-shell
    # doesn't work reliably when run via curl | bash)
    _ensure_path_in_profile "$HOME/.local/bin"
else
    info "$(msg '使用 pip 安装...' 'Installing with pip...')"
    if [ "$BRANCH" = "main" ]; then
        "$PYTHON_CMD" -m pip install --upgrade --no-cache-dir "git+${REPO}.git"
    else
        "$PYTHON_CMD" -m pip install --upgrade --no-cache-dir "git+${REPO}.git@${BRANCH}"
    fi
fi

# Post-install pip remnant cleanup
while IFS= read -r sp_dir; do
    [ -d "$sp_dir" ] || continue
    for remnant in "$sp_dir"/~*; do
        [ -e "$remnant" ] || continue
        rm -rf "$remnant" 2>/dev/null || true
    done
done < <("$PYTHON_CMD" -c "import site
for p in site.getsitepackages():
    print(p)" 2>/dev/null || true)

# ─── Step 5: Verify ───
step "$(msg '步骤 5/5: 验证安装' 'Step 5/5: Verifying installation')"

# Also check common install locations as fallback
if ! command -v agentflow >/dev/null 2>&1; then
    for _bin_dir in "$HOME/.local/bin" "$HOME/.cargo/bin" "/usr/local/bin"; do
        if [ -x "$_bin_dir/agentflow" ]; then
            export PATH="$_bin_dir:$PATH"
            break
        fi
    done
fi

if command -v agentflow >/dev/null 2>&1; then
    ok "$(msg 'agentflow 命令已就绪！' 'agentflow command is ready!')"
    agentflow version 2>/dev/null || true
else
    warn "$(msg 'agentflow 未在 PATH 中找到。' 'agentflow not found in PATH.')"
    warn "$(msg '请重启终端，或手动执行:' 'Please restart your terminal, or run:')"
    printf "\n${CYAN}    export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}\n\n"
fi

# ─── Launch interactive menu ───
printf "\n${BOLD}${GREEN}  ✅ $(msg '安装完成！' 'Installation complete!')${RESET}\n"
printf "${DIM}  $(msg '请重启终端或执行 source ~/.zshrc 后即可使用 agentflow 命令' 'Restart your terminal or run: source ~/.zshrc')${RESET}\n\n"

if command -v agentflow >/dev/null 2>&1; then
    printf "$(msg '  正在启动交互式菜单...' '  Launching interactive menu...')\n\n"
    agentflow </dev/tty
fi
