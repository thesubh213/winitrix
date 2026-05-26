@echo off
title Winitrix
setlocal
set "APP_DIR=%~dp0"
cd /d "%APP_DIR%" >nul 2>&1
"%APP_DIR%winitrix.exe" %*
set "EXIT_CODE=%ERRORLEVEL%"
endlocal & exit /b %EXIT_CODE%