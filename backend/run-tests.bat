@echo off
echo Running LearnVibe tests...
powershell -ExecutionPolicy Bypass -File "%~dp0scripts\run-tests.ps1"
if %ERRORLEVEL% NEQ 0 (
    echo Tests failed with exit code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
) else (
    echo Tests completed successfully
    exit /b 0
) 