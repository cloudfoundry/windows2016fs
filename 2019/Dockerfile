FROM mcr.microsoft.com/windowsservercore-insider:10.0.17744.1001

RUN cmd.exe /C net users /ADD vcap /passwordreq:no /expires:never && runas /user:vcap whoami
RUN cmd.exe /C net accounts /maxpwage:UNLIMITED

RUN powershell.exe -Command \
  $ErrorActionPreference = 'Stop'; \
  \
  Add-WindowsFeature Web-Webserver, \
    Web-WebSockets, \
    Web-WHC, \
    Web-ASP, \
    Web-ASP-Net45

COPY Git-*-64-bit.exe /git-setup.exe
RUN C:\git-setup.exe /SILENT /NORESTART
RUN del /F C:\git-setup.exe

COPY rewrite*.msi /Windows/rewrite.msi
RUN msiexec /i C:\Windows\rewrite.msi /qn /quiet

RUN powershell.exe -command \
	$sstfile = [System.IO.Path]::GetTempPath() + 'roots.sst'; \
	certutil -generateSSTFromWU $sstfile; \
	Import-Certificate -CertStoreLocation Cert:\LocalMachine\Root $sstfile; \
	Remove-Item -path $sstfile

RUN powershell.exe -command "start-bitstransfer https://aka.ms/vs/15/release/vc_redist.x64.exe $env:TEMP\vc_redist.x64.exe"
RUN cmd.exe /s /c "%TEMP%\vc_redist.x64.exe /install /passive /norestart /wait"
RUN del /F "%TEMP%\vc_redist.x64.exe"

RUN powershell.exe -command "remove-windowsfeature -name 'windows-defender'"

RUN powershell.exe -command \
  $svs=('AppHostSvc', 'MSDTC', 'TermService', 'WAS', 'dhcp', 'diagtrack', 'lmhosts', 'w3svc', 'winrm', 'RemoteRegistry'); \
  foreach ($name in $svs) { Set-Service -Name $name -StartupType Disabled }
