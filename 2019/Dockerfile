FROM mcr.microsoft.com/windowsservercore-insider:10.0.17733.1000

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

RUN powershell.exe -command "remove-windowsfeature -name 'windows-defender-features'"

RUN powershell.exe -command \
  $svs=('AppHostSvc', 'MSDTC', 'TermService', 'WAS', 'dhcp', 'diagtrack', 'lmhosts', 'w3svc', 'winrm', 'RemoteRegistry'); \
  foreach ($name in $svs) { Set-Service -Name $name -StartupType Disabled }