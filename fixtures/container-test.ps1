$ErrorActionPreference = "Stop";
trap { 
    $host.SetShouldExit(1) 
}
netsh interface portproxy add v4tov4 listenport=445 connectport=1025

net use t: $env:SHARE_UNC $env:SHARE_PASSWORD /user:$env:SHARE_USERNAME
if ($LASTEXITCODE -ne 0) {
    echo "ERROR: could not create smb mapping"
    Get-EventLog -LogName System -Newest 3 | format-list -Property Message

    exit $LASTEXITCODE
}

Start-Sleep 1

net use
if ($LASTEXITCODE -ne 0) {
    echo "ERROR: could not read smb mappings"
    Get-EventLog -LogName System -Newest 3 | format-list -Property Message

    exit $LASTEXITCODE
}

Start-Sleep 1
