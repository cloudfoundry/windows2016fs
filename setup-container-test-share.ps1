Param(
    [parameter(Mandatory=$true)]
    $action
)

if ($action -eq "remove") {
    Remove-LocalUser -Name $env:SHARE_USERNAME
    Remove-LocalUser -Name $env:SHARE_USERNAME2
    Remove-SmbShare -Name $env:SHARE_NAME -Force 
}

if ($action -eq "add") {
    $shareSourcePath = [System.IO.Path]::GetTempPath() 
    net user $env:SHARE_USERNAME $env:SHARE_PASSWORD /add
    net user $env:SHARE_USERNAME2 $env:SHARE_PASSWORD /add
    New-SmbShare -Name $env:SHARE_NAME -Path $shareSourcePath -ErrorAction Stop
}

Get-SmbShare | Format-Table
Get-LocalUser | Format-Table