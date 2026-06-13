@echo off
REM ============================================================================
REM Flowork Agent - one-click start (Windows).
REM Builds flowork-gui.exe on first run (needs Go 1.25+), seeds the bundled
REM agents + apps, then serves the control panel on http://127.0.0.1:1987
REM Override the address with FLOWORK_ADDR.
REM ============================================================================
setlocal enabledelayedexpansion
cd /d "%~dp0"

if not defined FLOWORK_ADDR set FLOWORK_ADDR=127.0.0.1:1987
if not defined FLOWORK_POWER_ARMED set FLOWORK_POWER_ARMED=1

REM --- Build the binary if missing (or run with an existing one). ---
if not exist "bin" mkdir bin
if not exist "bin\flowork-gui.exe" (
    echo - building flowork-gui ^(one-time, needs Go 1.25+^)...
    go build -o bin\flowork-gui.exe .
    if errorlevel 1 ( echo build failed. & pause & exit /b 1 )
)

REM --- Build the group-template wasm if missing (needed for Group create). ---
if exist "templates\group-template\main.go" if not exist "templates\group-template\agent.wasm" (
    echo - building group template wasm...
    pushd templates\group-template
    set "GOOS=wasip1" & set "GOARCH=wasm"
    go build -o agent.wasm .
    set "GOOS=" & set "GOARCH="
    popd
)

REM --- Seed bundled agents -> %USERPROFILE%\.flowork\agents (only if absent). ---
set "AG_DST=%USERPROFILE%\.flowork\agents"
if not exist "%AG_DST%" mkdir "%AG_DST%"
for /d %%d in (agents\*.fwagent) do (
    if not exist "%AG_DST%\%%~nxd" (
        xcopy /E /I /Q /Y "%%d" "%AG_DST%\%%~nxd" >nul
        if exist "%AG_DST%\%%~nxd\workspace" rmdir /S /Q "%AG_DST%\%%~nxd\workspace"
    )
)

REM --- Seed bundled apps -> %USERPROFILE%\.flowork\apps (only if absent). ---
set "APP_DST=%USERPROFILE%\.flowork\apps"
if not exist "%APP_DST%" mkdir "%APP_DST%"
for /d %%d in (apps\*) do (
    if not exist "%APP_DST%\%%~nxd" xcopy /E /I /Q /Y "%%d" "%APP_DST%\%%~nxd" >nul
)

echo Flowork Agent - starting on http://%FLOWORK_ADDR%  (power armed=%FLOWORK_POWER_ARMED%)
echo Schedules ^& triggers boot automatically inside the agent.
bin\flowork-gui.exe -addr "%FLOWORK_ADDR%"
pause
