# AgentFlow Installer for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# ─── Configuration ───
$REPO = if ($env:AGENTFLOW_REPO) { $env:AGENTFLOW_REPO } else { "https://github.com/kittors/AgentFlow" }
$BRANCH = if ($env:AGENTFLOW_BRANCH) { $env:AGENTFLOW_BRANCH } else { "main" }
$GITHUB_API = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"
$INSTALL_DIR = Join-Path $HOME ".agentflow" "bin"

# ─── Locale detection ───
$USE_ZH = $false
try {
    $culture = (Get-Culture).Name
    if ($culture -match '^zh') { $USE_ZH = $true }
} catch {}

function msg($zh, $en) {
    if ($USE_ZH) { return $zh } else { return $en }
}

function Write-Step($text)  { Write-Host "`n  --- $text ---`n" -ForegroundColor Cyan }
function Write-Ok($text)    { Write-Host "  [✓]  $text" -ForegroundColor Green }
function Write-Info($text)  { Write-Host "  [·]  $text" -ForegroundColor Cyan }
function Write-Warn($text)  { Write-Host "  [!]  $text" -ForegroundColor Yellow }
function Write-Err($text)   { Write-Host "  [✗]  $text" -ForegroundColor Red; exit 1 }

# ─── Safe file removal (handles Windows file locking) ───
function Safe-Remove($path) {
    if (Test-Path $path) {
        try {
            Remove-Item -Recurse -Force $path -ErrorAction Stop
        } catch {
            # Rename-aside fallback for locked files
            $aside = "$path._agentflow_old_$(Get-Date -Format 'yyyyMMddHHmmss')"
            try {
                Rename-Item $path $aside -ErrorAction Stop
                Write-Warn (msg "文件被锁定，已重命名: $aside" "File locked, renamed aside: $aside")
            } catch {
                Write-Warn (msg "无法移除: $path" "Cannot remove: $path")
            }
        }
    }
}

# ─── Banner ───
Write-Host ""
Write-Host "     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗" -ForegroundColor Cyan
Write-Host "    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║" -ForegroundColor Cyan
Write-Host "    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║" -ForegroundColor Cyan
Write-Host "    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║" -ForegroundColor Cyan
Write-Host "    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝" -ForegroundColor Cyan
Write-Host "    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝ " -ForegroundColor Cyan
Write-Host "    Multi-CLI Agent Workflow System" -ForegroundColor DarkGray
Write-Host ""

# ─── Step 1: Detect Git ───
Write-Step (msg "步骤 1/5: 检测 Git" "Step 1/5: Detecting Git")

try {
    $gitVersion = & git --version 2>&1
    Write-Ok (msg "找到 Git ($gitVersion)" "Found Git ($gitVersion)")
} catch {
    Write-Err (msg "Git 未找到。请先安装 Git。" "Git not found. Please install Git first.")
}

# ─── Step 2: Detect Python ───
Write-Step (msg "步骤 2/5: 检测 Python" "Step 2/5: Detecting Python")

$PYTHON_CMD = $null
$INSTALL_MODE = "binary"   # default to binary, upgrade to pip if Python found

foreach ($cmd in @("python3", "python", "py")) {
    try {
        $ver = & $cmd --version 2>&1
        if ($ver -match '(\d+)\.(\d+)') {
            $major = [int]$Matches[1]
            $minor = [int]$Matches[2]
            if ($major -gt 3 -or ($major -eq 3 -and $minor -ge 10)) {
                $PYTHON_CMD = $cmd
                $INSTALL_MODE = "pip"
                break
            }
        }
    } catch {}
}

if ($INSTALL_MODE -eq "pip") {
    $pyVer = & $PYTHON_CMD --version 2>&1
    Write-Ok (msg "找到 $PYTHON_CMD ($pyVer)" "Found $PYTHON_CMD ($pyVer)")
} else {
    Write-Warn (msg "未找到 Python >= 3.10，将下载预编译的独立版本。" "Python >= 3.10 not found. Will download pre-built standalone binary.")
}

# ─── Step 3: Detect package manager ───
Write-Step (msg "步骤 3/5: 检测包管理器" "Step 3/5: Detecting package manager")

$HAS_UV = $false
if ($INSTALL_MODE -eq "pip") {
    try {
        $uvVer = & uv --version 2>&1
        $HAS_UV = $true
        Write-Ok (msg "找到 uv ($uvVer)，将优先使用" "Found uv ($uvVer), will use it")
    } catch {
        Write-Info (msg "未找到 uv，将使用 pip" "uv not found, falling back to pip")
    }
} else {
    Write-Info (msg "独立二进制模式，跳过包管理器检测" "Standalone binary mode, skipping package manager detection")
}

# ─── Step 4: Install ───
Write-Step (msg "步骤 4/5: 安装 AgentFlow (分支: $BRANCH)" "Step 4/5: Installing AgentFlow (branch: $BRANCH)")

if ($INSTALL_MODE -eq "pip") {
    # ── Pip/uv install path ──
    if ($HAS_UV) {
        Write-Info (msg "使用 uv 安装..." "Installing with uv...")
        if ($BRANCH -eq "main") {
            & uv tool install --force --from "git+$REPO" agentflow
        } else {
            & uv tool install --force --from "git+$REPO@$BRANCH" agentflow
        }
    } else {
        Write-Info (msg "使用 pip 安装..." "Installing with pip...")
        if ($BRANCH -eq "main") {
            & $PYTHON_CMD -m pip install --upgrade --no-cache-dir "git+$REPO.git"
        } else {
            & $PYTHON_CMD -m pip install --upgrade --no-cache-dir "git+$REPO.git@$BRANCH"
        }
    }
} else {
    # ── Binary download fallback ──
    Write-Info (msg "正在从 GitHub Releases 下载预编译版本..." "Downloading pre-built binary from GitHub Releases...")

    try {
        $releaseInfo = Invoke-RestMethod -Uri $GITHUB_API -Headers @{ "User-Agent" = "AgentFlow-Installer" }
        $asset = $releaseInfo.assets | Where-Object { $_.name -like "agentflow-windows-amd64*" } | Select-Object -First 1

        if (-not $asset) {
            Write-Err (msg "未在最新 Release 中找到 Windows 二进制文件。请先安装 Python >= 3.10 后重试。" "No Windows binary found in latest release. Please install Python >= 3.10 and try again.")
        }

        $downloadUrl = $asset.browser_download_url
        Write-Info (msg "下载: $($asset.name) ..." "Downloading: $($asset.name) ...")

        # Create install directory
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }

        $exePath = Join-Path $INSTALL_DIR "agentflow.exe"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing

        Write-Ok (msg "已下载到 $exePath" "Downloaded to $exePath")

        # Add to PATH (user-level, persistent)
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$INSTALL_DIR*") {
            [Environment]::SetEnvironmentVariable("Path", "$INSTALL_DIR;$currentPath", "User")
            $env:Path = "$INSTALL_DIR;$env:Path"
            Write-Ok (msg "已将 $INSTALL_DIR 添加到用户 PATH" "Added $INSTALL_DIR to user PATH")
        }
    } catch {
        Write-Err (msg "下载失败: $_`n请安装 Python >= 3.10 后再试: https://www.python.org/downloads/" "Download failed: $_`nPlease install Python >= 3.10 and try again: https://www.python.org/downloads/")
    }
}

# ─── Step 5: Verify ───
Write-Step (msg "步骤 5/5: 验证安装" "Step 5/5: Verifying installation")

try {
    & agentflow version 2>&1
    Write-Ok (msg "agentflow 命令已就绪！" "agentflow command is ready!")
} catch {
    Write-Warn (msg "agentflow 未在 PATH 中找到。" "agentflow not found in PATH.")
    Write-Warn (msg "可能需要重启终端或将安装路径加入 PATH。" "You may need to restart your terminal or add the install location to PATH.")
}

# ─── Launch interactive menu ───
Write-Host ""
Write-Host (msg "  ✅ 安装完成！正在启动交互式菜单..." "  ✅ Installation complete! Launching interactive menu...") -ForegroundColor Green
Write-Host ""

try {
    & agentflow
} catch {}
