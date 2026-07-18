@echo off
REM build.bat — two-stage build for Language Betawi + its installer.
REM
REM STAGE 1 must run first: it compiles the actual language compiler
REM (betawi.exe) and copies it into internal/assets/bin/betawi.exe, which
REM STAGE 2 then embeds into betawi-installer.exe via go:embed. If you
REM skip stage 1, the installer will "successfully" extract a useless
REM 1-byte placeholder file.

echo === Stage 1: building betawi.exe (the language compiler) ===
go build -o betawi.exe .\cmd\betawi
if errorlevel 1 goto :error

echo === Copying betawi.exe into internal\assets\bin for embedding ===
copy /Y betawi.exe internal\assets\bin\betawi.exe
if errorlevel 1 goto :error

echo === Embedding the manifest (admin elevation + Common Controls v6) ===
echo === into betawi-installer.exe as a real Windows resource ===
where rsrc >nul 2>nul
if errorlevel 1 (
    echo rsrc tool not found - installing it once via 'go install'...
    go install github.com/akavel/rsrc@latest
    if errorlevel 1 goto :error
)
rsrc -manifest betawi-installer.exe.manifest -o cmd\betawi-installer\rsrc.syso
if errorlevel 1 goto :error

echo === Stage 2: building betawi-installer.exe (embeds betawi.exe + manifest above) ===
REM -H=windowsgui marks this as a GUI-subsystem binary, so Windows does NOT
REM spawn a console host window behind it (that's what caused the wide
REM black terminal-looking frame around the wizard). betawi.exe itself
REM stays a normal console app since it needs a real terminal for running
REM .bwi scripts.
go build -ldflags "-H=windowsgui" -o betawi-installer.exe .\cmd\betawi-installer
if errorlevel 1 goto :error

echo.
echo Done. Two files were produced:
echo   betawi.exe            - the language compiler itself (run .bwi scripts directly)
echo   betawi-installer.exe  - the setup wizard end users should actually download
echo.
echo betawi-installer.exe now has the manifest EMBEDDED (real Windows resource,
echo via rsrc.syso), so it will prompt for admin elevation automatically on
echo double-click. The old sidecar betawi-installer.exe.manifest file is no
echo longer needed for that — safe to delete, or just leave it, it's harmless.
goto :eof

:error
echo.
echo Build failed - see the error above.
exit /b 1
