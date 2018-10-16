$ErrorActionPreference = 'Stop'
trap { $host.SetShouldExit(1) }

$acl = Get-Acl HKLM:\SOFTWARE\Microsoft\SystemCertificates\ROOT
echo $acl
$rule = $acl.Access | ?{ $_.IdentityReference -like "*\vcap" }
echo $rule
$hasrule = $rule.RegistryRights -like "*CreateSubKey*"
echo $hasrule
If ($hasrule -eq $False) {
	echo "Vcap user does not have permission to add custom certificate for LocalMachine/Root"
	exit 1
}
