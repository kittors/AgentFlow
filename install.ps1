# AgentFlow Installer for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$GITHUB_API = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"
$INSTALL_DIR = Join-Path (Join-Path $HOME ".agentflow") "bin"
$PREVIOUS_AGENTFLOW = $null
try {
    $existing = Get-Command agentflow -ErrorAction SilentlyContinue
    if ($existing) { $PREVIOUS_AGENTFLOW = $existing.Source }
} catch {}

$USE_ZH = $false
try {
    if ((Get-Culture).Name -match '^zh') { $USE_ZH = $true }
} catch {}

function msg($zh, $en) {
    if ($USE_ZH) { return $zh }
    return $en
}

function Write-Step($text)  { Write-Host "`n  --- $text ---`n" -ForegroundColor Cyan }
function Write-Ok($text)    { Write-Host "  [✓]  $text" -ForegroundColor Green }
function Write-Info($text)  { Write-Host "  [·]  $text" -ForegroundColor Cyan }
function Write-Err($text)   { Write-Host "  [✗]  $text" -ForegroundColor Red; exit 1 }

function Resolve-AssetName {
    $arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "X64" { "amd64" }
        "Arm64" { "arm64" }
        default { Write-Err (msg "不支持的 CPU 架构" "Unsupported CPU architecture") }
    }
    return "agentflow-windows-$arch.exe"
}

function Resolve-DownloadUrl {
    if ($env:AGENTFLOW_DOWNLOAD_URL) {
        return $env:AGENTFLOW_DOWNLOAD_URL
    }

    if ($env:AGENTFLOW_BRANCH -and $env:AGENTFLOW_BRANCH -ne "main") {
        Write-Err (msg "当前安装器仅支持发布版本；自定义分支请在本地构建 Go 二进制。" "This installer supports released builds only; for a custom branch, build the Go binary locally.")
    }

    $assetName = Resolve-AssetName
    $releaseInfo = Invoke-RestMethod -Uri $GITHUB_API -Headers @{ "User-Agent" = "AgentFlow-Installer" }
    $asset = $releaseInfo.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1
    if (-not $asset) {
        Write-Err (msg "未在最新 Release 中找到匹配当前平台的二进制文件。" "No matching binary was found in the latest release.")
    }
    return $asset.browser_download_url
}

function Show-ShadowWarning {
    if ($PREVIOUS_AGENTFLOW -and ($PREVIOUS_AGENTFLOW -ne (Join-Path $INSTALL_DIR "agentflow.exe"))) {
        Write-Host ("  [!]  " + (msg "检测到旧的 agentflow 可能仍在当前终端中抢先命中: $PREVIOUS_AGENTFLOW" "An older agentflow may still shadow the new binary in your current shell: $PREVIOUS_AGENTFLOW")) -ForegroundColor Yellow
        Write-Host ("       " + (msg "请重新打开终端，或运行: `$env:Path = `"$INSTALL_DIR;`$env:Path`"" "Reopen your terminal, or run: `$env:Path = `"$INSTALL_DIR;`$env:Path`"")) -ForegroundColor Yellow
    }
}

function Start-AgentFlowMenu {
    if ($env:AGENTFLOW_NO_TUI -eq "1") {
        return
    }

    try {
        if (-not [Environment]::UserInteractive) {
            return
        }
    } catch {
        return
    }

    Write-Info (msg "首次启动 AgentFlow..." "Launching AgentFlow for the first time...")
    try {
        & $exePath
    } catch {
        Write-Host ("  [!]  " + (msg "未能自动进入 AgentFlow 菜单，请稍后手动运行 agentflow。" "AgentFlow could not be started automatically; run agentflow manually.")) -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "AgentFlow" -ForegroundColor Cyan
Write-Host (msg "Go Binary Installer" "Go Binary Installer") -ForegroundColor DarkGray
Write-Host ""

Write-Step (msg "步骤 1/3: 解析下载地址" "Step 1/3: Resolve download URL")
$downloadUrl = Resolve-DownloadUrl
Write-Ok (msg "已获取下载地址" "Download URL resolved")

Write-Step (msg "步骤 2/3: 下载 AgentFlow" "Step 2/3: Download AgentFlow")
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
}
$exePath = Join-Path $INSTALL_DIR "agentflow.exe"
Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing
Write-Ok (msg "已下载到 $exePath" "Downloaded to $exePath")

Write-Step (msg "步骤 3/3: 配置 PATH 并验证" "Step 3/3: Configure PATH and verify")
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$INSTALL_DIR*") {
    [Environment]::SetEnvironmentVariable("Path", "$INSTALL_DIR;$currentPath", "User")
}
$env:Path = "$INSTALL_DIR;$env:Path"

& $exePath version
Show-ShadowWarning
Write-Host ""
Write-Host (msg "  ✅ 安装完成" "  ✅ Installation complete") -ForegroundColor Green
Start-AgentFlowMenu
