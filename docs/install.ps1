$ErrorActionPreference = "Stop"
$installDir = Join-Path $env:LOCALAPPDATA "paymo-cli"
$zipPath = Join-Path $env:TEMP "paymo-cli.zip"
$url = "https://github.com/mbundgaard/paymo-cli/releases/latest/download/paymo-cli_windows_amd64.zip"

Write-Host "Installing paymo-cli to $installDir..."

New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Invoke-WebRequest -Uri $url -OutFile $zipPath
tar -xf $zipPath -C $installDir
Remove-Item $zipPath -ErrorAction SilentlyContinue

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added $installDir to PATH."
}

$version = & (Join-Path $installDir "paymo.exe") --version 2>&1
Write-Host "$version installed successfully."
Write-Host "Run 'paymo auth login' to get started."
