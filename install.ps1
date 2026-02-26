# AgentFlow Installer for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# ─── Configuration ───
$REPO = "https://github.com/kittors/AgentFlow"
$BRANCH = if ($env:AGENTFLOW_BRANCH) { $env:AGENTFLOW_BRANCH } else { "main" }

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
foreach ($cmd in @("python3", "python", "py")) {
    try {
        $ver = & $cmd --version 2>&1
        if ($ver -match '(\d+)\.(\d+)') {
            $major = [int]$Matches[1]
            $minor = [int]$Matches[2]
            if ($major -gt 3 -or ($major -eq 3 -and $minor -ge 10)) {
                $PYTHON_CMD = $cmd
                break
            }
        }
    } catch {}
}

if (-not $PYTHON_CMD) {
    Write-Err (msg "需要 Python >= 3.10，但未找到。" "Python >= 3.10 is required but not found.")
}
$pyVer = & $PYTHON_CMD --version 2>&1
Write-Ok (msg "找到 $PYTHON_CMD ($pyVer)" "Found $PYTHON_CMD ($pyVer)")

# ─── Step 3: Detect package manager ───
Write-Step (msg "步骤 3/5: 检测包管理器" "Step 3/5: Detecting package manager")

$HAS_UV = $false
try {
    $uvVer = & uv --version 2>&1
    $HAS_UV = $true
    Write-Ok (msg "找到 uv ($uvVer)，将优先使用" "Found uv ($uvVer), will use it")
} catch {
    Write-Info (msg "未找到 uv，将使用 pip" "uv not found, falling back to pip")
}

# ─── Step 4: Install ───
Write-Step (msg "步骤 4/5: 安装 AgentFlow (分支: $BRANCH)" "Step 4/5: Installing AgentFlow (branch: $BRANCH)")

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
