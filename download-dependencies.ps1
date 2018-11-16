param (
    [string]$dependenciesDir  
)

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$ProgressPreference = 'SilentlyContinue'

curl -UseBasicParsing -Outfile $dependenciesDir/Git-VERSION-64-bit.exe -Uri "https://github.com/git-for-windows/git/releases/download/v2.20.0-rc0.windows.1/Git-2.20.0.rc0.windows.1-64-bit.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/tar-VERSION.exe        -Uri "https://s3.amazonaws.com/bosh-windows-dependencies/tar-1536096948.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/rewrite.msi            -Uri "https://download.microsoft.com/download/C/9/E/C9E8180D-4E51-40A6-A9BF-776990D8BCA9/rewrite_amd64.msi"
curl -UseBasicParsing -Outfile $dependenciesDir/vc_redist.x64.exe      -Uri "https://aka.ms/vs/15/release/vc_redist.x64.exe"