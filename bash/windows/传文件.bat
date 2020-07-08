@echo off
setlocal enabledelayedexpansion

color 0A
set key=xxx
for /f "tokens=1,2 delims==" %%i in (config.ini) do (
    REM echo %%i
    if %%i==%key% (
        set value=%%j
        REM echo !value!
    )
    
)
echo filename is: %value%
install.exe %value%
pause