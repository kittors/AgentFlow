@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

echo.
echo   ╔══════════════════════════════════════╗
echo   ║       AgentFlow Installer            ║
echo   ╚══════════════════════════════════════╝
echo.

REM ── Detect architecture ──
set "ARCH="
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="x86" set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"
if "%ARCH%"=="" (
    echo   [X]  不支持的 CPU 架构: %PROCESSOR_ARCHITECTURE%
    echo        Unsupported CPU architecture: %PROCESSOR_ARCHITECTURE%
    pause
    exit /b 1
)

REM ── Locate binary in the same directory as this script ──
set "SCRIPT_DIR=%~dp0"
set "SRC_BIN="

REM Try exact name first
if exist "%SCRIPT_DIR%agentflow.exe" (
    set "SRC_BIN=%SCRIPT_DIR%agentflow.exe"
    goto :found
)

REM Try architecture-specific name
for %%F in ("%SCRIPT_DIR%agentflow-windows-*.exe") do (
    set "SRC_BIN=%%F"
    goto :found
)

echo   [X]  未找到 agentflow 可执行文件
echo        No agentflow executable found in: %SCRIPT_DIR%
echo        请确保 install.bat 和 agentflow.exe 在同一目录下。
pause
exit /b 1

:found
echo   [1/3]  检测到二进制文件: %SRC_BIN%

REM ── Create install directory ──
set "INSTALL_DIR=%USERPROFILE%\.agentflow\bin"
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM ── Copy binary ──
echo   [2/3]  复制到 %INSTALL_DIR%\agentflow.exe ...
copy /Y "%SRC_BIN%" "%INSTALL_DIR%\agentflow.exe" >nul
if errorlevel 1 (
    echo   [X]  复制失败，请检查权限
    pause
    exit /b 1
)

REM ── Add to PATH ──
echo   [3/3]  配置 PATH 环境变量...

REM Check if already in PATH
echo %PATH% | findstr /I /C:".agentflow\bin" >nul 2>&1
if %errorlevel%==0 (
    echo   [ok]  PATH 中已存在 .agentflow\bin
    goto :verify
)

REM Add to user PATH via PowerShell (most reliable way on modern Windows)
powershell -NoProfile -Command ^
    "$p = [Environment]::GetEnvironmentVariable('Path','User'); ^
     if ($p -notlike '*\.agentflow\bin*') { ^
         [Environment]::SetEnvironmentVariable('Path', '%INSTALL_DIR%;' + $p, 'User') ^
     }"

REM Update current session PATH
set "PATH=%INSTALL_DIR%;%PATH%"

:verify
echo.
echo   ──────────────────────────────────────
"%INSTALL_DIR%\agentflow.exe" version
echo   ──────────────────────────────────────
echo.
echo   ✅ 安装完成！
echo   Installation complete!
echo.
echo   请重新打开终端窗口，然后运行: agentflow
echo   Please reopen your terminal, then run: agentflow
echo.
pause
