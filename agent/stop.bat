@echo off
REM Stop the Flowork agent (Windows).
taskkill /F /IM flowork-gui.exe >nul 2>&1 && (echo Agent stopped.) || (echo Agent was not running.)
