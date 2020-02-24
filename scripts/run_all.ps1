param (
    [string]$Command="ginkgo",
    [Parameter(Mandatory=$true)]
    [switch]$ConfirmTheStemcellsAreUpToDate
)

$ErrorActionPreference = "Stop"

Write-Output "Running 2019"
. .\.envrc-2019.ps1
Powershell $Command
