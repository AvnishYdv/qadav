@echo off
setlocal
echo ================================================================================
echo   QADAV (.ydv) - ONE-CLICK WINDOWS SYSTEM INSTALLER
echo ================================================================================
echo.
powershell -ExecutionPolicy Bypass -File "%~dp0scripts\install_qadav.ps1" %*
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Installation failed with exit code %ERRORLEVEL%.
    pause
    exit /b %ERRORLEVEL%
)
echo.
echo [SUCCESS] Qadav is now installed and ready to use!
pause
