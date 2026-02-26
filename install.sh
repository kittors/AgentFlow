#!/usr/bin/env bash
set -eu

# ─── Configuration ───
REPO="https://github.com/kittors/AgentFlow"
BRANCH="${AGENTFLOW_BRANCH:-main}"

# ─── Colors ───
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

# ─── Locale detection ───
USE_ZH=false
_locale="${LC_ALL:-${LC_MESSAGES:-${LANG:-${LANGUAGE:-}}}}"
case "$_locale" in
    zh*|ZH*) USE_ZH=true ;;
esac

msg() {
    if [ "$USE_ZH" = true ]; then echo "$1"; else echo "$2"; fi
}

info()    { printf "${CYAN}  [·]${RESET}  %s\n" "$*"; }
ok()      { printf "${GREEN}  [✓]${RESET}  %s\n" "$*"; }
warn()    { printf "${YELLOW}  [!]${RESET}  %s\n" "$*"; }
err()     { printf "${RED}  [✗]${RESET}  %s\n" "$*"; exit 1; }
step()    { printf "\n${BOLD}${CYAN}  ─── %s ───${RESET}\n\n" "$*"; }

# ─── Banner ───
printf "\n"
printf "${BOLD}${CYAN}"
printf "     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗\n"
printf "    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║\n"
printf "    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║\n"
printf "    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║\n"
printf "    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝\n"
printf "    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝ \n"
printf "${RESET}"
printf "${DIM}    Multi-CLI Agent Workflow System${RESET}\n"
printf "\n"

# ─── Step 1: Detect Git ───
step "$(msg '步骤 1/5: 检测 Git' 'Step 1/5: Detecting Git')"

if ! command -v git >/dev/null 2>&1; then
    err "$(msg 'Git 未找到。请先安装 Git。' 'Git not found. Please install Git first.')"
fi
ok "$(msg "找到 Git ($(git --version 2>&1))" "Found Git ($(git --version 2>&1))")"

# ─── Step 2: Detect Python ───
step "$(msg '步骤 2/5: 检测 Python' 'Step 2/5: Detecting Python')"

PYTHON_CMD=""
for cmd in python3 python; do
    if command -v "$cmd" >/dev/null 2>&1; then
        version=$("$cmd" --version 2>&1 | grep -oE '[0-9]+\.[0-9]+' | head -1)
        major=$(echo "$version" | cut -d. -f1)
        minor=$(echo "$version" | cut -d. -f2)
        if [ "$major" -gt 3 ] || { [ "$major" -eq 3 ] && [ "$minor" -ge 10 ]; }; then
            PYTHON_CMD="$cmd"
            break
        fi
    fi
done

if [ -z "$PYTHON_CMD" ]; then
    err "$(msg '需要 Python >= 3.10，但未找到。' 'Python >= 3.10 is required but not found.')"
fi
ok "$(msg "找到 $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))" "Found $PYTHON_CMD ($($PYTHON_CMD --version 2>&1))")"

# ─── Step 3: Detect package manager ───
step "$(msg '步骤 3/5: 检测包管理器' 'Step 3/5: Detecting package manager')"

HAS_UV=false
if command -v uv >/dev/null 2>&1; then
    HAS_UV=true
    ok "$(msg "找到 uv ($(uv --version 2>&1))，将优先使用" "Found uv ($(uv --version 2>&1)), will use it")"
else
    info "$(msg '未找到 uv，将使用 pip' 'uv not found, falling back to pip')"
fi

# ─── Step 3.5: Clean pip remnants ───
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

# ─── Step 4: Install ───
step "$(msg "步骤 4/5: 安装 AgentFlow (分支: $BRANCH)" "Step 4/5: Installing AgentFlow (branch: $BRANCH)")"

if [ "$HAS_UV" = true ]; then
    info "$(msg '使用 uv 安装...' 'Installing with uv...')"
    if [ "$BRANCH" = "main" ]; then
        uv tool install --force --from "git+${REPO}" agentflow
    else
        uv tool install --force --from "git+${REPO}@${BRANCH}" agentflow
    fi
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

if command -v agentflow >/dev/null 2>&1; then
    ok "$(msg 'agentflow 命令已就绪！' 'agentflow command is ready!')"
    agentflow version 2>/dev/null || true
else
    warn "$(msg 'agentflow 未在 PATH 中找到。' 'agentflow not found in PATH.')"
    warn "$(msg '可能需要重启终端或将安装路径加入 PATH。' 'You may need to restart your terminal or add the install location to PATH.')"
fi

# ─── Launch interactive menu ───
printf "\n${BOLD}${GREEN}  ✅ $(msg '安装完成！正在启动交互式菜单...' 'Installation complete! Launching interactive menu...')${RESET}\n\n"

if command -v agentflow >/dev/null 2>&1; then
    agentflow </dev/tty
fi
