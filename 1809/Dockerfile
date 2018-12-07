FROM mcr.microsoft.com/windows/servercore:1809

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

COPY vc_redist.x64.exe /vc_redist.x64.exe
RUN cmd.exe /s /c "c:\vc_redist.x64.exe /install /passive /norestart /wait"
RUN del /F "c:\vc_redist.x64.exe"

RUN powershell.exe -command "remove-windowsfeature -name 'windows-defender'"

# enable on-demand start of DNS cache to support FQDNs for net use
RUN powershell.exe -command "Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\dnscache' -Name Start -Value 3"

# enable on-demand start of lmhosts service
RUN powershell.exe -command "Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\lmhosts' -Name Start -Value 3"

RUN powershell.exe -command \
  $svs=('AppHostSvc', 'MSDTC', 'TermService', 'WAS', 'dhcp', 'diagtrack', 'w3svc', 'winrm', 'RemoteRegistry'); \
  foreach ($name in $svs) { Set-Service -Name $name -StartupType Disabled }