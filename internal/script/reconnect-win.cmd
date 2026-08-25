REM {{ if .FileName }}{{.FileName}}{{ else }}gws-ssh-with-reconnect-{{.Name}}.cmd{{ end }}
@echo off
setlocal

REM This script automatically reconnects to an SSH server when the connection drops.
REM It continuously attempts to establish an SSH connection and retries after a specified delay if disconnected.

REM The number of seconds to wait before retrying the connection.
set "RETRY_SECONDS=3"

REM Prepare escape sequence for zellij reset
for /F %%a in ('echo prompt $E^|cmd') do set "ESC=%%a"

:reconnect_loop
ssh {{.User}}@{{.Host}} -p {{.Port}}{{ if .KnownHostsFile }} -o UserKnownHostsFile={{.KnownHostsFile}}{{ end }}
set "SSH_EXIT=%errorlevel%"
<nul set /p "=%ESC%[?1000l%ESC%[?1002l%ESC%[?1003l%ESC%[?1006l%ESC%c"
if %SSH_EXIT% == 0 goto end
cls
echo Disconnected. Reconnecting in %RETRY_SECONDS% seconds...
timeout /t %RETRY_SECONDS% /nobreak
cls
goto reconnect_loop

:end
