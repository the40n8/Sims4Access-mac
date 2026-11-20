# SimSight Companion for Windows in one command (PowerShell, the game closed):
#
#   irm https://github.com/the40n8/simsight-companion/releases/latest/download/install.ps1 | iex
#
# Downloads the release, checks its SHA-256 fingerprint, then runs the
# installer inside it. Run it again to update, or to remove SimSight.
# Published as install.ps1 next to the release zip (windows/get.ps1 in the
# source). Env for testing: SIMSIGHT_REPO, SIMSIGHT_BASE.
#
# Everything runs inside one script block, called on the last line: a
# download cut short runs nothing. It never calls exit, which would close
# the PowerShell window the player typed into. ASCII only (Windows
# PowerShell 5.1).
& {
  $ErrorActionPreference = 'Stop'
  $ProgressPreference = 'SilentlyContinue'   # progress bars are noise for screen readers
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

  function Tell([string]$Text) { Write-Host $Text }

  $repo = 'the40n8/simsight-companion'
  if ($env:SIMSIGHT_REPO) { $repo = $env:SIMSIGHT_REPO }
  $base = "https://github.com/$repo/releases/latest/download"
  if ($env:SIMSIGHT_BASE) { $base = $env:SIMSIGHT_BASE.TrimEnd('/') }
  $zip = 'SimSight-Companion-windows.zip'

  Tell 'SimSight Companion, accessibility mod for The Sims 4.'
  if (Get-Process -Name 'TS4_x64', 'TS4_DX9_x64', 'TS4' -ErrorAction SilentlyContinue) {
    Tell 'The Sims 4 is running. Save, quit the game, then run this command again. Nothing was changed.'
    return
  }

  $tmp = Join-Path ([IO.Path]::GetTempPath()) ('simsight-' + [guid]::NewGuid().ToString('N'))
  New-Item -ItemType Directory -Path $tmp | Out-Null
  try {
    Tell 'Downloading.'
    try {
      Invoke-WebRequest -UseBasicParsing -Uri "$base/$zip" -OutFile (Join-Path $tmp $zip)
      Invoke-WebRequest -UseBasicParsing -Uri "$base/SHA256SUMS" -OutFile (Join-Path $tmp 'SHA256SUMS')
    } catch {
      Tell 'Download failed. Check the internet connection and try again. Nothing was changed.'
      return
    }
    $want = ''
    foreach ($line in Get-Content -LiteralPath (Join-Path $tmp 'SHA256SUMS')) {
      $parts = $line -split '\s+', 2
      if ($parts.Count -eq 2 -and $parts[1].TrimStart('*') -eq $zip) { $want = $parts[0].ToLower() }
    }
    $have = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $tmp $zip)).Hash.ToLower()
    if (-not $want -or $want -ne $have) {
      Tell 'The download does not match its fingerprint, so it was not installed. Try again later.'
      return
    }
    Tell 'Download checked.'
    Expand-Archive -LiteralPath (Join-Path $tmp $zip) -DestinationPath (Join-Path $tmp 'x')
    $dir = Join-Path $tmp 'x\SimSight-Companion'
    $installer = Join-Path $dir 'install.ps1'
    if (-not (Test-Path -LiteralPath $installer)) {
      Tell 'The download is incomplete. Nothing was changed.'
      return
    }
    $version = Join-Path $dir 'VERSION'
    if (Test-Path -LiteralPath $version) {
      foreach ($line in Get-Content -LiteralPath $version) { if ($line -like 'release=*') { Tell ('Version: ' + $line.Substring(8)) } }
    }
    # Its own PowerShell, allowed to run this one script whatever the
    # computer's script policy; it asks before changing anything.
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $installer
  } finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
  }
}
