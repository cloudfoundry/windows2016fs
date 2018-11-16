$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }


net use t: \\$env:SHARE_HOST\my-share $env:SHARE_PASSWORD /user:$env:SHARE_USERNAME
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}