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
       irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1 | iex

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

Remote examples:
  irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1 | iex
  & ([scriptblock]::Create((irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1))) -Yes
  & ([scriptblock]::Create((irm https://github.com/natsuz0ra/SlimeBot/releases/latest/download/uninstall.ps1))) -Purge
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

function Split-PathEntries {
  param([string]$Value)
  if ([string]::IsNullOrWhiteSpace($Value)) {
    return @()
  }
  return @($Value -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

function Normalize-PathEntry {
  param([string]$Value)
  return $Value.Trim().TrimEnd('\', '/')
}

function Remove-PathEntry {
  param(
    [string]$Value,
    [string]$Target
  )

  $normalizedTarget = Normalize-PathEntry $Target
  $entries = @(Split-PathEntries $Value)
  $kept = @()
  foreach ($entry in $entries) {
    if (-not [string]::Equals((Normalize-PathEntry $entry), $normalizedTarget, [System.StringComparison]::OrdinalIgnoreCase)) {
      $kept += $entry
    }
  }
  return ($kept -join ';')
}

function Remove-UserPathEntry {
  param([string]$Directory)

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $updatedUserPath = Remove-PathEntry $userPath $Directory
  if ($updatedUserPath -ne $userPath) {
    [Environment]::SetEnvironmentVariable("Path", $updatedUserPath, "User")
    Write-Host "Removed command shim directory from user PATH: $Directory"
  }

  $env:Path = Remove-PathEntry $env:Path $Directory
}

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
Remove-UserPathEntry $shimDir

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
Write-Host "Open a new terminal to refresh PATH."
Write-Host "SlimeBot uninstall complete"
