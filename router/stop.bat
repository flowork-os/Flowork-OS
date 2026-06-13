@echo off
REM Flow Router — one-click stop (Windows)
echo Flow Router — stopping...
taskkill /F /IM flow-router-bin.exe >nul 2>&1
if errorlevel 1 (
    echo (not running)
) else (
    echo Stopped.
)

REM One-click teardown: unload local model from VRAM (free GPU). Best-effort.
if not defined FLOWORK_LOCAL_MODEL set FLOWORK_LOCAL_MODEL=qwen-flowork
where ollama >nul 2>&1 && (
    ollama stop %FLOWORK_LOCAL_MODEL% >nul 2>&1
    echo Local model %FLOWORK_LOCAL_MODEL% unloaded.
)
pause
