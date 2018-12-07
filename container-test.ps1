$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters'

net use t: $env:SHARE_UNC $env:SHARE_PASSWORD /user:$env:SHARE_USERNAME
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Start-Sleep 1