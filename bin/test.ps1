$ProgressPreference="SilentlyContinue"
$ErrorActionPreference = "Stop";
$installerLink = "";
$patchLink = "";
$packagePath = "";
trap { $host.SetShouldExit(1) }

function Run-Docker {
  param([String[]] $cmd)

  docker @cmd
  if ($LASTEXITCODE -ne 0) {
    Exit $LASTEXITCODE
  }
}

function Get-Dotnet-Paths {
  Write-Host "Extracting dotnet paths from DOTNET_DOCKERFILE..."
  # Powershell variable scopes are...interesting
  $script:installerLink=Get-Content $env:DOTNET_DOCKERFILE | Select-String -Pattern 'https.*\b' | Select-String -Pattern 'dotnet-framework-installer' | Select -Expand Matches | Select -ExpandProperty Value
  $script:patchLink=Get-Content $env:DOTNET_DOCKERFILE | Select-String -Pattern 'https.*\b' | Select-String -Pattern 'patch\.msu' | Select -Expand Matches | Select -ExpandProperty Value
  $script:packagePath=Get-Content $env:DOTNET_DOCKERFILE | Select-String -Pattern '/PackagePath.*\b' | Select-String -Pattern 'dism' | Select -Expand Matches | Select -ExpandProperty Value

  Remove-Item ./dotnet-dockerfile

  $pathProblem=$false
  if ([string]::IsNullOrWhitespace($installerLink)) {
    Write-Host "Installer Link Not Found"
    $pathProblem=$true
  }
  if ([string]::IsNullOrWhitespace($patchLink)) {
    Write-Host "Patch Link Not Found"
    $pathProblem=$true
  }
  if ([string]::IsNullOrWhitespace($packagePath)) {
    Write-Host "Package Path Not Found"
    $pathProblem=$true
  }

  Write-Host "Found installer link: $installerLink"
  Write-Host "Found patch link: $patchLink"
  Write-Host "Found package path: $packagePath"

  if ($pathProblem) {
    Exit 1
  }
}

restart-service docker

$version=(cat $env:VERSION_NUMBER)
$digest=(cat $env:UPSTREAM_IMAGE_DIGEST)

Push-Location "$env:BUILT_BINARIES"

Get-Dotnet-Paths

Write-Host "Building image using the '$digest' provided by Concourse"
Run-Docker "--version"
# Pre-written debug just in case
# Write-Host "Running docker build with these parameters: BASE_IMAGE_DIGEST=@$digest, DOTNET_INSTALLER_LINK=$installerLink, DOTNET_PATCH_LINK=$patchLink, DOTNET_PACKAGE_PATH=$packagePath"
Run-Docker "build", "--build-arg", "BASE_IMAGE_DIGEST=@$digest", "--build-arg", "DOTNET_INSTALLER_LINK=$installerLink", "--build-arg", "DOTNET_PATCH_LINK=$patchLink", "--build-arg", "DOTNET_PACKAGE_PATH=$packagePath", "-t", "$env:IMAGE_NAME", "-t", "${env:IMAGE_NAME}:$version", "-t", "${env:IMAGE_NAME}:${env:OS_VERSION}", "--pull", "."

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
