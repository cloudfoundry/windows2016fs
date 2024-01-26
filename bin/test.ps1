$ProgressPreference="SilentlyContinue"
$ErrorActionPreference = "Stop";
trap { $host.SetShouldExit(1) }

function Run-Docker {
  param([String[]] $cmd)

  docker @cmd
  if ($LASTEXITCODE -ne 0) {
    Exit $LASTEXITCODE
  }
}

restart-service docker

$version=(cat $env:VERSION_NUMBER)
$digest=(cat $env:UPSTREAM_IMAGE_DIGEST)

Push-Location "$env:BUILT_BINARIES"

Write-Host "Building image using the '$digest' provided by Concourse"
Run-Docker "--version"
Run-Docker "build", "--build-arg", "BASE_IMAGE_DIGEST=@$digest", "-t", "$env:IMAGE_NAME", "-t", "${env:IMAGE_NAME}:$version", "-t", "${env:IMAGE_NAME}:${env:OS_VERSION}", "--pull", "."

# output systeminfo including hotfixes for documentation
Run-Docker "run", "${env:IMAGE_NAME}:$version", "cmd", "/c", "systeminfo"
Run-Docker "run", "${env:IMAGE_NAME}:$version", "powershell", "(get-childitem C:\Windows\System32\msvcr100.dll).VersionInfo | Select-Object -Property FileDescription,ProductVersion"
Run-Docker "run", "${env:IMAGE_NAME}:$version", "powershell", "(get-childitem C:\Windows\System32\vcruntime140.dll).VersionInfo | Select-Object -Property FileDescription,ProductVersion"

$env:TEST_CANDIDATE_IMAGE=$env:IMAGE_NAME
$env:VERSION_TAG=$env:OS_VERSION
Pop-Location

Invoke-Expression "go run github.com/onsi/ginkgo/v2/ginkgo $args"
if ($LASTEXITCODE -ne 0) {
  throw "tests failed"
}

Run-Docker "images", "-a"
Run-Docker "login", "-u", "$env:DOCKER_USERNAME", "-p", "$env:DOCKER_PASSWORD"
Run-Docker "push", "${env:IMAGE_NAME}:latest"
Run-Docker "push", "${env:IMAGE_NAME}:$version"
Run-Docker "push", "${env:IMAGE_NAME}:${env:OS_VERSION}"
