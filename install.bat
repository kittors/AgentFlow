@echo off
chcp 65001 >nul 2>&1
setlocal

echo.
echo   ========================================
echo        AgentFlow Installer
echo   ========================================
echo.

REM ── Detect architecture ──
set "ARCH="
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="x86" set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"
if "%ARCH%"=="" (
    echo   [X]  Unsupported CPU: %PROCESSOR_ARCHITECTURE%
    goto :end
)
echo   [ok]  CPU: %PROCESSOR_ARCHITECTURE%

REM ── Locate binary ──
set "SCRIPT_DIR=%~dp0"
set "SRC_BIN="

if exist "%SCRIPT_DIR%agentflow.exe" (
    set "SRC_BIN=%SCRIPT_DIR%agentflow.exe"
    goto :found
)

for %%F in ("%SCRIPT_DIR%agentflow-windows-*.exe") do (
    set "SRC_BIN=%%F"
    goto :found
)

echo   [X]  agentflow.exe not found in: %SCRIPT_DIR%
echo        Please make sure install.bat and agentflow.exe are in the same folder.
goto :end

:found
echo   [ok]  Found: %SRC_BIN%

REM ── Create install directory ──
set "INSTALL_DIR=%USERPROFILE%\.agentflow\bin"
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM ── Copy binary ──
echo   [..]  Copying to %INSTALL_DIR%\agentflow.exe ...
copy /Y "%SRC_BIN%" "%INSTALL_DIR%\agentflow.exe" >nul
if errorlevel 1 (
    echo   [X]  Copy failed. Check permissions.
    goto :end
)
echo   [ok]  Copied successfully.

REM ── Add to PATH ──
echo   [..]  Configuring PATH...

echo %PATH% | findstr /I /C:".agentflow\bin" >nul 2>&1
if %errorlevel%==0 (
    echo   [ok]  PATH already configured.
    goto :verify
)

powershell -NoProfile -Command "$p = [Environment]::GetEnvironmentVariable('Path','User'); if ($p -notlike '*.agentflow\bin*') { [Environment]::SetEnvironmentVariable('Path', '%INSTALL_DIR%;' + $p, 'User') }"

set "PATH=%INSTALL_DIR%;%PATH%"
echo   [ok]  PATH updated.

:verify
echo.
echo   ----------------------------------------
"%INSTALL_DIR%\agentflow.exe" version
echo   ----------------------------------------
echo.
echo   Installation complete!
echo   Please reopen your terminal, then run: agentflow
echo.

:end
pause
