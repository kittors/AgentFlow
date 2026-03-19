#!/usr/bin/env bash
# AgentFlow Local Installer
# This script installs from the same directory (no network required).
# Usage: Unpack the archive, then run: bash install.sh

set -euo pipefail

# ── Colors ──
RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; NC='\033[0m'
step()  { printf "\n${CYAN}  --- %s ---${NC}\n\n" "$1"; }
ok()    { printf "${GREEN}  [✓]  %s${NC}\n" "$1"; }
info()  { printf "${CYAN}  [·]  %s${NC}\n" "$1"; }
err()   { printf "${RED}  [✗]  %s${NC}\n" "$1"; exit 1; }

# ── Language ──
USE_ZH=false
if [[ "${LANG:-}" =~ ^zh ]] || [[ "${LC_ALL:-}" =~ ^zh ]]; then
    USE_ZH=true
fi

msg() {
    if $USE_ZH; then echo "$1"; else echo "$2"; fi
}

printf "\n${CYAN}AgentFlow Installer${NC}\n\n"

# ── Locate binary ──
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_BIN=""

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux*)  GOOS="linux" ;;
    darwin*) GOOS="darwin" ;;
    *)       err "$(msg "不支持的操作系统: $OS" "Unsupported OS: $OS")" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  GOARCH="amd64" ;;
    aarch64|arm64) GOARCH="arm64" ;;
    *)             err "$(msg "不支持的 CPU 架构: $ARCH" "Unsupported architecture: $ARCH")" ;;
esac

# Try to find binary
EXPECTED_NAME="agentflow-${GOOS}-${GOARCH}"
for candidate in \
    "${SCRIPT_DIR}/agentflow" \
    "${SCRIPT_DIR}/${EXPECTED_NAME}" \
    "${SCRIPT_DIR}"/agentflow-*; do
    if [[ -f "$candidate" && ! "$candidate" =~ \.(sh|md|txt|bat)$ ]]; then
        SRC_BIN="$candidate"
        break
    fi
done

if [[ -z "$SRC_BIN" ]]; then
    err "$(msg "未找到 agentflow 可执行文件，请确保 install.sh 和二进制文件在同一目录" \
               "No agentflow binary found. Ensure install.sh and the binary are in the same directory.")"
fi

step "$(msg "步骤 1/3: 检测环境" "Step 1/3: Detect environment")"
info "$(msg "系统: ${GOOS}/${GOARCH}" "System: ${GOOS}/${GOARCH}")"
ok "$(msg "已找到二进制文件: $(basename "$SRC_BIN")" "Found binary: $(basename "$SRC_BIN")")"

# ── Install ──
step "$(msg "步骤 2/3: 安装 AgentFlow" "Step 2/3: Install AgentFlow")"
INSTALL_DIR="$HOME/.agentflow/bin"
mkdir -p "$INSTALL_DIR"

cp "$SRC_BIN" "$INSTALL_DIR/agentflow"
chmod +x "$INSTALL_DIR/agentflow"
ok "$(msg "已安装到 $INSTALL_DIR/agentflow" "Installed to $INSTALL_DIR/agentflow")"

# ── macOS: remove quarantine ──
if [[ "$GOOS" == "darwin" ]]; then
    xattr -d com.apple.quarantine "$INSTALL_DIR/agentflow" 2>/dev/null || true
fi

# ── Configure PATH ──
step "$(msg "步骤 3/3: 配置 PATH" "Step 3/3: Configure PATH")"

add_to_path() {
    local line="export PATH=\"\$HOME/.agentflow/bin:\$PATH\""
    local rc_file="$1"
    if [[ -f "$rc_file" ]]; then
        if ! grep -qF '.agentflow/bin' "$rc_file" 2>/dev/null; then
            printf '\n# AgentFlow\n%s\n' "$line" >> "$rc_file"
            ok "$(msg "已添加到 $rc_file" "Added to $rc_file")"
        else
            ok "$(msg "$rc_file 中已存在 PATH 配置" "PATH already configured in $rc_file")"
        fi
    fi
}

PATH_ADDED=false
CURRENT_SHELL="$(basename "${SHELL:-bash}")"
case "$CURRENT_SHELL" in
    zsh)
        add_to_path "$HOME/.zshrc"
        PATH_ADDED=true
        ;;
    bash)
        if [[ "$GOOS" == "darwin" ]]; then
            add_to_path "$HOME/.bash_profile"
        else
            add_to_path "$HOME/.bashrc"
        fi
        PATH_ADDED=true
        ;;
    fish)
        FISH_CONFIG="$HOME/.config/fish/config.fish"
        if [[ -f "$FISH_CONFIG" ]]; then
            if ! grep -qF '.agentflow/bin' "$FISH_CONFIG" 2>/dev/null; then
                echo 'set -gx PATH $HOME/.agentflow/bin $PATH' >> "$FISH_CONFIG"
                ok "$(msg "已添加到 $FISH_CONFIG" "Added to $FISH_CONFIG")"
            fi
        fi
        PATH_ADDED=true
        ;;
esac

# Also try common rc files if main shell config wasn't found
if ! $PATH_ADDED; then
    for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [[ -f "$rc" ]]; then
            add_to_path "$rc"
            PATH_ADDED=true
            break
        fi
    done
fi

export PATH="$INSTALL_DIR:$PATH"

# ── Verify ──
echo ""
echo "  ──────────────────────────────────────"
"$INSTALL_DIR/agentflow" version
echo "  ──────────────────────────────────────"
echo ""
printf "${GREEN}  ✅ $(msg "安装完成！" "Installation complete!")${NC}\n"
echo ""
echo "  $(msg "请重新打开终端，然后运行: agentflow" "Reopen your terminal, then run: agentflow")"
echo ""
