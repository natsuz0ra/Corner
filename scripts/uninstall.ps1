param(
  [switch]$Yes,
  [switch]$Purge,
  [switch]$KeepData,
  [switch]$Help
)

$ErrorActionPreference = "Stop"

function Show-Usage {
  Write-Host @"
Usage: .\uninstall.ps1 [-Yes] [-Purge] [-KeepData] [-Help]

Uninstalls SlimeBot service, installed files, and command shims.

Options:
  -Yes       Run non-interactively and keep user data.
  -Purge     Run non-interactively and delete user data (~\.slimebot by default).
  -KeepData  Keep user data without prompting.
  -Help      Show this help.

Environment:
  SLIMEBOT_INSTALL_DIR  Install directory (default: %LOCALAPPDATA%\SlimeBot)
  SLIMEBOT_BIN_DIR      Command shim directory (default: install dir\bin)
  SLIMEBOT_HOME         User data directory (default: %USERPROFILE%\.slimebot)
"@
}

if ($Help) {
  Show-Usage
  exit 0
}

if ($Purge -and $KeepData) {
  Write-Error "-Purge and -KeepData cannot be used together"
  exit 2
}

$installDir = if ($env:SLIMEBOT_INSTALL_DIR) { $env:SLIMEBOT_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "SlimeBot" }
$shimDir = if ($env:SLIMEBOT_BIN_DIR) { $env:SLIMEBOT_BIN_DIR } else { Join-Path $installDir "bin" }
$dataDir = if ($env:SLIMEBOT_HOME) { $env:SLIMEBOT_HOME } else { Join-Path $env:USERPROFILE ".slimebot" }

$slimebotPath = Join-Path $installDir "bin\slimebot.exe"
if (-not (Test-Path $slimebotPath)) {
  $command = Get-Command slimebot -ErrorAction SilentlyContinue
  if ($command) {
    $slimebotPath = $command.Source
  } else {
    $slimebotPath = $null
  }
}

function Invoke-ServiceAction {
  param([string]$Action)

  if (-not $slimebotPath) {
    Write-Host "Skipping service $Action`: slimebot command not found"
    return
  }

  try {
    & $slimebotPath service $Action
    if ($LASTEXITCODE -eq 0) {
      Write-Host "Service $Action completed"
    } else {
      Write-Host "Service $Action failed or was not needed; continuing"
    }
  } catch {
    Write-Host "Service $Action failed or was not needed; continuing"
  }
}

Invoke-ServiceAction "stop"
Invoke-ServiceAction "uninstall"

Remove-Item -Force -ErrorAction SilentlyContinue (Join-Path $shimDir "slimebot.cmd"), (Join-Path $shimDir "slimebot-cli.cmd")
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $installDir

$deleteData = $false
if ($Purge) {
  $deleteData = $true
} elseif ($KeepData -or $Yes) {
  $deleteData = $false
} elseif (Test-Path $dataDir) {
  $answer = Read-Host "Delete SlimeBot user data at $dataDir? This removes config, database, uploads, and skills. [y/N]"
  if ($answer -match '^(y|yes)$') {
    $deleteData = $true
  }
}

if ($deleteData) {
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $dataDir
  Write-Host "Deleted user data: $dataDir"
} else {
  Write-Host "Kept user data: $dataDir"
}

Write-Host "Removed install directory: $installDir"
Write-Host "Removed command shims from: $shimDir"
Write-Host "SlimeBot uninstall complete"
