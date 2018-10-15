$ErrorActionPreference = 'Stop'
trap { $host.SetShouldExit(1) }

$username = "vcap"
$acl = Get-Acl HKLM:\SOFTWARE\Microsoft\SystemCertificates\ROOT

# remove existing permissions, if any
$rule = $acl.Access | ?{ $_.IdentityReference -like "*\$username" }
if ($rule) {
    $rule | %{ $acl.RemoveAccessRule($_) | Out-Null }
}

# create new minimum acl allowing vcap to import a root CA cert to
# the local machine store
$rule = New-Object System.Security.AccessControl.RegistryAccessRule `
    $username, `
    @("CreateSubKey, ReadKey, SetValue, Delete"), `
    @("ContainerInherit, ObjectInherit"), `
    "None", `
    "Allow"

$acl.AddAccessRule($rule)
$acl | Set-Acl
