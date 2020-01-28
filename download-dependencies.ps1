param (
    [Parameter(Mandatory=$True)]
    [string]$dependenciesDir  
)

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$ProgressPreference = 'SilentlyContinue'

curl -UseBasicParsing -Outfile $dependenciesDir/Git-VERSION-64-bit.exe   -Uri "https://github.com/git-for-windows/git/releases/download/v2.20.0-rc0.windows.1/Git-2.20.0.rc0.windows.1-64-bit.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/tar-VERSION.exe          -Uri "https://s3.amazonaws.com/bosh-windows-dependencies/tar-1536096948.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/rewrite.msi              -Uri "https://download.microsoft.com/download/C/9/E/C9E8180D-4E51-40A6-A9BF-776990D8BCA9/rewrite_amd64.msi"
curl -UseBasicParsing -Outfile $dependenciesDir/vcredist-2010.x64.exe    -Uri "https://download.microsoft.com/download/1/6/5/165255E7-1014-4D0A-B094-B6A430A6BFFC/vcredist_x64.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/vcredist-ucrt.x64.exe    -Uri "https://aka.ms/vs/16/release/vc_redist.x64.exe"
curl -UseBasicParsing -Outfile $dependenciesDir/dotnet-48-installer.exe  -Uri "https://download.visualstudio.microsoft.com/download/pr/014120d7-d689-4305-befd-3cb711108212/0fd66638cde16859462a6243a4629a50/ndp48-x86-x64-allos-enu.exe"