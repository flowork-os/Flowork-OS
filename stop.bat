@echo off
REM Flowork - stop the whole stack (Windows): agent + router.
echo Flowork - stopping...
taskkill /F /IM flowork-gui.exe     >nul 2>&1 && (echo - Agent stopped.)  || (echo - Agent was not running.)
taskkill /F /IM flow-router-bin.exe >nul 2>&1 && (echo - Router stopped.) || (echo - Router was not running.)
echo Done.
pause
