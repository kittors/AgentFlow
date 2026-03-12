#!/usr/bin/env bash
set -eu

REPO="${AGENTFLOW_REPO:-https://github.com/kittors/AgentFlow}"
GITHUB_TAG_API="https://api.github.com/repos/kittors/AgentFlow/releases/tags/continuous"
GITHUB_LATEST_API="https://api.github.com/repos/kittors/AgentFlow/releases/latest"
INSTALL_DIR="${HOME}/.agentflow/bin"
BRANCH="${AGENTFLOW_BRANCH:-main}"
PREVIOUS_AGENTFLOW="$(command -v agentflow 2>/dev/null || true)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

USE_ZH=false
locale_value="${LC_ALL:-${LC_MESSAGES:-${LANG:-${LANGUAGE:-}}}}"
case "${locale_value}" in
    zh*|ZH*) USE_ZH=true ;;
esac

msg() {
    if [ "${USE_ZH}" = true ]; then echo "$1"; else echo "$2"; fi
}

info()  { printf "${CYAN}  [·]${RESET}  %s\n" "$*"; }
ok()    { printf "${GREEN}  [✓]${RESET}  %s\n" "$*"; }
warn()  { printf "${YELLOW}  [!]${RESET}  %s\n" "$*"; }
err()   { printf "${RED}  [✗]${RESET}  %s\n" "$*"; exit 1; }
step()  { printf "\n${BOLD}${CYAN}  ─── %s ───${RESET}\n\n" "$*"; }

ensure_line_in_profile() {
    search_key="$1"
    export_line="$2"
    label="$3"
    current_shell="${SHELL:-}"
    case "${current_shell}" in
        */zsh) profiles="${HOME}/.zshrc" ;;
        */bash) profiles="${HOME}/.bashrc ${HOME}/.bash_profile" ;;
        *) profiles="${HOME}/.profile" ;;
    esac

    updated=false
    for profile in ${profiles}; do
        [ -f "${profile}" ] || touch "${profile}"
        if ! grep -qF "${search_key}" "${profile}" 2>/dev/null; then
            printf "\n# Added by AgentFlow installer\n%s\n" "${export_line}" >> "${profile}"
            updated=true
            ok "$(msg "${label} 已写入 $(basename "${profile}")" "${label} added to $(basename "${profile}")")"
        fi
    done
    if [ "${updated}" = false ]; then
        ok "$(msg "${label} 已存在于 shell 配置中" "${label} already configured in shell profile")"
    fi
}

persist_path() {
    ensure_line_in_profile "${INSTALL_DIR}" "export PATH=\"${INSTALL_DIR}:\$PATH\"" "PATH"
}

print_shell_refresh_notice() {
    if [ -n "${PREVIOUS_AGENTFLOW}" ] && [ "${PREVIOUS_AGENTFLOW}" != "${INSTALL_DIR}/agentflow" ]; then
        warn "$(msg "检测到旧的 agentflow 仍可能在当前终端中抢先命中: ${PREVIOUS_AGENTFLOW}" "An older agentflow may still shadow the new binary in your current shell: ${PREVIOUS_AGENTFLOW}")"
        printf "     %s\n" "$(msg "运行: export PATH=\"${INSTALL_DIR}:\$PATH\" && hash -r" "Run: export PATH=\"${INSTALL_DIR}:\$PATH\" && hash -r")"
        printf "     %s\n" "$(msg "或重新打开终端 / source 对应 shell 配置文件" "Or reopen your terminal / source your shell profile")"
    fi
}

can_launch_tui() {
    [ "${AGENTFLOW_NO_TUI:-0}" != "1" ] && [ -r /dev/tty ] && [ -w /dev/tty ]
}

launch_tui() {
    if ! can_launch_tui; then
        return 0
    fi

    info "$(msg "首次启动 AgentFlow..." "Launching AgentFlow for the first time...")"
    if ! "${INSTALL_DIR}/agentflow" </dev/tty >/dev/tty 2>/dev/tty; then
        warn "$(msg "未能自动进入 AgentFlow 菜单，请稍后手动运行 agentflow。" "AgentFlow could not be started automatically; run agentflow manually.")"
    fi
}

resolve_platform() {
    os_name="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch_name="$(uname -m)"

    case "${os_name}" in
        linux*) platform="linux" ;;
        darwin*) platform="darwin" ;;
        *) err "$(msg "不支持的操作系统: ${os_name}" "Unsupported OS: ${os_name}")" ;;
    esac

    case "${arch_name}" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) err "$(msg "不支持的架构: ${arch_name}" "Unsupported architecture: ${arch_name}")" ;;
    esac

    ASSET_NAME="agentflow-${platform}-${arch}"
}

release_json_from_api() {
    endpoint="$1"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -H "User-Agent: AgentFlow-Installer" "${endpoint}"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- --header="User-Agent: AgentFlow-Installer" "${endpoint}"
    else
        err "$(msg "需要 curl 或 wget 来下载 AgentFlow" "curl or wget is required to download AgentFlow")"
    fi
}

download_url_from_release_api() {
    endpoint="$1"
    release_json="$(release_json_from_api "${endpoint}")" || return 1

    printf "%s" "${release_json}" | grep -o "\"browser_download_url\"[[:space:]]*:[[:space:]]*\"[^\"]*${ASSET_NAME}[^\"]*\"" \
        | head -1 \
        | sed 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
}

download_url_from_preferred_release() {
    download_url="$(download_url_from_release_api "${GITHUB_TAG_API}" || true)"
    if [ -n "${download_url}" ]; then
        printf "%s" "${download_url}"
        return 0
    fi

    download_url="$(download_url_from_release_api "${GITHUB_LATEST_API}" || true)"
    if [ -n "${download_url}" ]; then
        printf "%s" "${download_url}"
        return 0
    fi

    return 1
}

download_binary() {
    resolve_platform

    if [ -n "${AGENTFLOW_DOWNLOAD_URL:-}" ]; then
        download_url="${AGENTFLOW_DOWNLOAD_URL}"
    else
        if [ "${BRANCH}" != "main" ]; then
            err "$(msg "二进制安装器当前仅支持发布版本；自定义分支请先用 Go 在本地构建。" "The binary installer currently supports released builds only; for a custom branch, build locally with Go first.")"
        fi
        download_url="$(download_url_from_preferred_release)" || err "$(msg "无法获取最新版本下载地址" "Failed to resolve latest release download URL")"
    fi

    [ -n "${download_url}" ] || err "$(msg "未找到匹配当前平台的二进制文件" "No matching binary found for this platform")"

    mkdir -p "${INSTALL_DIR}"
    binary_path="${INSTALL_DIR}/agentflow"

    info "$(msg "下载 ${ASSET_NAME} ..." "Downloading ${ASSET_NAME} ...")"
    if command -v curl >/dev/null 2>&1; then
        curl -fL -o "${binary_path}" "${download_url}" || err "$(msg "下载失败" "Download failed")"
    else
        wget -qO "${binary_path}" "${download_url}" || err "$(msg "下载失败" "Download failed")"
    fi
    chmod +x "${binary_path}"
}

printf "\n"
printf "${BOLD}${CYAN}AgentFlow${RESET}\n"
printf "${DIM}%s${RESET}\n\n" "$(msg "Go Binary Installer" "Go Binary Installer")"

step "$(msg "步骤 1/3: 检测下载工具" "Step 1/3: Detect download tools")"
if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
    err "$(msg "未找到 curl 或 wget" "curl or wget was not found")"
fi
ok "$(msg "下载工具可用" "Download tool available")"

step "$(msg "步骤 2/3: 下载 AgentFlow Go 二进制" "Step 2/3: Download AgentFlow Go binary")"
download_binary
persist_path
ok "$(msg "已安装到 ${INSTALL_DIR}/agentflow" "Installed to ${INSTALL_DIR}/agentflow")"

step "$(msg "步骤 3/3: 验证安装" "Step 3/3: Verify installation")"
PATH="${INSTALL_DIR}:$PATH"
if [ -x "${INSTALL_DIR}/agentflow" ]; then
    ok "$(msg "agentflow 命令已就绪" "agentflow command is ready")"
    "${INSTALL_DIR}/agentflow" version || true
    print_shell_refresh_notice
else
    warn "$(msg "agentflow 还未进入当前 PATH，请重开终端后再试。" "agentflow is not yet on your current PATH; reopen the terminal and try again.")"
fi

printf "\n${BOLD}${GREEN}  ✅ %s${RESET}\n" "$(msg "安装完成" "Installation complete")"
launch_tui
