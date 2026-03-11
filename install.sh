#!/usr/bin/env bash
set -eu

# ─── Configuration ───
REPO="${AGENTFLOW_REPO:-https://github.com/kittors/AgentFlow}"
BRANCH="${AGENTFLOW_BRANCH:-main}"
GITHUB_API="https://api.github.com/repos/kittors/AgentFlow/releases/latest"
INSTALL_DIR="$HOME/.agentflow/bin"

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

# ─── Ensure env var in shell profile ───
_ensure_env_in_profile() {
    # Usage: _ensure_env_in_profile "SEARCH_KEY" "export LINE" "label"
    local search_key="$1"
    local export_line="$2"
    local label="$3"
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
        [ -f "$profile" ] || touch "$profile"
        if ! grep -qF "$search_key" "$profile" 2>/dev/null; then
            printf "\n# Added by AgentFlow installer\n%s\n" "$export_line" >> "$profile"
            updated=true
            ok "$(msg "$label 已写入 $(basename $profile)" "$label added to $(basename $profile)")"
        else
            # Update existing value if different
            local old_line
            old_line=$(grep -F "$search_key" "$profile" | tail -1)
            if [ "$old_line" != "$export_line" ]; then
                # Use sed to replace the old line
                sed -i'' -e "s|.*${search_key}.*|${export_line}|" "$profile"
                updated=true
                ok "$(msg "$label 已更新于 $(basename $profile)" "$label updated in $(basename $profile)")"
            fi
        fi
    done

    if [ "$updated" = false ]; then
        ok "$(msg "$label 已存在于 shell 配置中" "$label already configured in shell profile")"
    fi
}

_ensure_path_in_profile() {
    _ensure_env_in_profile "$1" "export PATH=\"${1}:\$PATH\"" "PATH"
}

_persist_lang_choice() {
    local lang="en"
    [ "$USE_ZH" = true ] && lang="zh"
    _ensure_env_in_profile "AGENTFLOW_LANG" "export AGENTFLOW_LANG=\"${lang}\"" "AGENTFLOW_LANG"
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

INSTALL_MODE="binary"   # default to binary, upgrade to pip if Python found
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
                INSTALL_MODE="pip"
                break
            fi
        fi
    fi
done

if [ "$INSTALL_MODE" = "pip" ]; then
    ok "$(msg "找到 $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))" "Found $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))")"
    # uv 有自己的 Python 管理，也走 pip 模式
    if [ -z "$PYTHON_CMD" ] && [ "$HAS_UV" = true ]; then
        warn "$(msg '未找到 Python >= 3.10，但已安装 uv。将使用 uv 继续安装。' 'Python >= 3.10 not found, but uv is installed. Proceeding with uv.')"
    fi
elif [ "$HAS_UV" = true ]; then
    # uv 可以自动管理 Python，走 pip 模式
    INSTALL_MODE="pip"
    warn "$(msg '未找到 Python >= 3.10，但已安装 uv。将使用 uv 继续安装。' 'Python >= 3.10 not found, but uv is installed. Proceeding with uv.')"
else
    warn "$(msg '未找到 Python >= 3.10，将下载预编译的独立版本。' 'Python >= 3.10 not found. Will download pre-built standalone binary.')"
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

if [ "$INSTALL_MODE" = "pip" ]; then
    # ── Pip/uv install path ──
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
        _persist_lang_choice
    else
        info "$(msg '使用 pip 安装...' 'Installing with pip...')"
        if [ "$BRANCH" = "main" ]; then
            "$PYTHON_CMD" -m pip install --upgrade --no-cache-dir "git+${REPO}.git"
        else
            "$PYTHON_CMD" -m pip install --upgrade --no-cache-dir "git+${REPO}.git@${BRANCH}"
        fi
        _persist_lang_choice
    fi

    # Post-install pip remnant cleanup
    if [ -n "$PYTHON_CMD" ]; then
        while IFS= read -r sp_dir; do
            [ -d "$sp_dir" ] || continue
            for remnant in "$sp_dir"/~*; do
                [ -e "$remnant" ] || continue
                rm -rf "$remnant" 2>/dev/null || true
            done
        done < <("$PYTHON_CMD" -c "import site
for p in site.getsitepackages():
    print(p)" 2>/dev/null || true)
    fi
else
    # ── Binary download fallback ──
    info "$(msg '正在从 GitHub Releases 下载预编译版本...' 'Downloading pre-built binary from GitHub Releases...')"

    # Detect OS and architecture
    _os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    _arch="$(uname -m)"
    case "$_os" in
        linux*)  _platform="linux" ;;
        darwin*) _platform="darwin" ;;
        *)       err "$(msg "不支持的操作系统: $_os" "Unsupported OS: $_os")" ;;
    esac
    case "$_arch" in
        x86_64|amd64) _arch_suffix="amd64" ;;
        aarch64|arm64) _arch_suffix="arm64" ;;
        *)            _arch_suffix="amd64" ;;
    esac
    _binary_name="agentflow-${_platform}-${_arch_suffix}"

    # Get download URL from GitHub API
    if command -v curl >/dev/null 2>&1; then
        _release_json=$(curl -fsSL -H "User-Agent: AgentFlow-Installer" "$GITHUB_API" 2>/dev/null) || \
            err "$(msg '无法获取 Release 信息。请检查网络并重试。' 'Failed to fetch release info. Check your network and try again.')"
    elif command -v wget >/dev/null 2>&1; then
        _release_json=$(wget -qO- --header="User-Agent: AgentFlow-Installer" "$GITHUB_API" 2>/dev/null) || \
            err "$(msg '无法获取 Release 信息。请检查网络并重试。' 'Failed to fetch release info. Check your network and try again.')"
    else
        err "$(msg '需要 curl 或 wget 来下载文件。' 'curl or wget is required to download files.')"
    fi

    # Parse download URL (use grep/sed for portability, no jq dependency)
    _download_url=$(echo "$_release_json" | grep -o '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]*'"$_binary_name"'[^"]*"' | head -1 | sed 's/.*"\(http[^"]*\)".*/\1/' | sed 's/"browser_download_url"[[:space:]]*:[[:space:]]*"//; s/"$//')

    if [ -z "$_download_url" ]; then
        err "$(msg "未在最新 Release 中找到 $_binary_name。请安装 Python >= 3.10 后重试。" "$_binary_name not found in latest release. Please install Python >= 3.10 and try again.")"
    fi

    info "$(msg "下载: $_binary_name ..." "Downloading: $_binary_name ...")"

    # Create install directory and download
    mkdir -p "$INSTALL_DIR"
    _exe_path="$INSTALL_DIR/agentflow"

    if command -v curl >/dev/null 2>&1; then
        curl -fSL -o "$_exe_path" "$_download_url" || err "$(msg '下载失败' 'Download failed')"
    else
        wget -qO "$_exe_path" "$_download_url" || err "$(msg '下载失败' 'Download failed')"
    fi
    chmod +x "$_exe_path"

    ok "$(msg "已下载到 $_exe_path" "Downloaded to $_exe_path")"

    # Add to PATH
    export PATH="$INSTALL_DIR:$PATH"
    _ensure_path_in_profile "$INSTALL_DIR"
    _persist_lang_choice
fi

# ─── Step 5: Verify ───
step "$(msg '步骤 5/5: 验证安装' 'Step 5/5: Verifying installation')"

# Also check common install locations as fallback
if ! command -v agentflow >/dev/null 2>&1; then
    for _bin_dir in "$INSTALL_DIR" "$HOME/.local/bin" "$HOME/.cargo/bin" "/usr/local/bin"; do
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
    # Pass language choice to the Python CLI
    if [ "$USE_ZH" = true ]; then
        AGENTFLOW_LANG=zh agentflow </dev/tty
    else
        AGENTFLOW_LANG=en agentflow </dev/tty
    fi
fi
