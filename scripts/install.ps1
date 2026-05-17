$ErrorActionPreference = "Stop"

$sourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$installDir = if ($env:SLIMEBOT_INSTALL_DIR) { $env:SLIMEBOT_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "SlimeBot" }
$shimDir = if ($env:SLIMEBOT_BIN_DIR) { $env:SLIMEBOT_BIN_DIR } else { Join-Path $installDir "bin" }

New-Item -ItemType Directory -Force -Path $installDir, $shimDir | Out-Null

Remove-Item -Recurse -Force -ErrorAction SilentlyContinue (Join-Path $installDir "bin"), (Join-Path $installDir "cli")
Copy-Item -Recurse -Force (Join-Path $sourceDir "bin") $installDir
Copy-Item -Recurse -Force (Join-Path $sourceDir "cli") $installDir
if (Test-Path (Join-Path $sourceDir "docs")) {
  Copy-Item -Recurse -Force (Join-Path $sourceDir "docs") $installDir
}
Copy-Item -Force (Join-Path $sourceDir "README.md"), (Join-Path $sourceDir "README.zh-CN.md"), (Join-Path $sourceDir "LICENSE") $installDir

$slimebotCmd = Join-Path $shimDir "slimebot.cmd"
$slimebotCliCmd = Join-Path $shimDir "slimebot-cli.cmd"
Set-Content -Path $slimebotCmd -Encoding ASCII -Value "@echo off`r`n`"$installDir\bin\slimebot.exe`" %*`r`n"
Set-Content -Path $slimebotCliCmd -Encoding ASCII -Value "@echo off`r`n`"$installDir\bin\slimebot.exe`" cli %*`r`n"

$homeDir = Join-Path $env:USERPROFILE ".slimebot"
New-Item -ItemType Directory -Force -Path $homeDir | Out-Null
$configPath = Join-Path $homeDir "config.cfg"
if (-not (Test-Path $configPath)) {
  Set-Content -Path $configPath -Encoding ASCII -Value @"
SERVER_PORT=6247
FRONTEND_PORT=7391
DB_PATH=~/.slimebot/storage/data.db
SKILLS_ROOT=~/.slimebot/skills
CHAT_UPLOAD_ROOT=~/.slimebot/storage/chat_uploads
WEB_SEARCH_API_KEY=YOUR_TAVILY_API_KEY
JWT_SECRET=CHANGE_ME_TO_A_RANDOM_SECRET
JWT_EXPIRE=21600
"@
}

Write-Host "SlimeBot installed to $installDir"
Write-Host "Command shims installed to $shimDir"
Write-Host "Run: slimebot"
Write-Host "If your shell cannot find slimebot, add this directory to PATH: $shimDir"
Write-Host "Before starting the web service, set JWT_SECRET in $configPath"
