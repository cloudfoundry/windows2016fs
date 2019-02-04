FROM microsoft/windowsservercore:1709

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

COPY tar-*.exe /Windows/tar.exe

COPY rewrite*.msi /Windows/rewrite.msi
RUN msiexec /i C:\Windows\rewrite.msi /qn /quiet

COPY vc_redist.x64.exe /vc_redist.x64.exe
RUN cmd.exe /s /c "c:\vc_redist.x64.exe /install /passive /norestart /wait"
RUN del /F "c:\vc_redist.x64.exe"

RUN powershell.exe -command "Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\dnscache' -Name Start -Value 4"

RUN powershell.exe -command "remove-windowsfeature -name 'windows-defender-features'"

RUN powershell.exe -command \
  $svs=('AppHostSvc', 'MSDTC', 'TermService', 'WAS', 'dhcp', 'diagtrack', 'w3svc', 'winrm', 'RemoteRegistry'); \
  foreach ($name in $svs) { Set-Service -Name $name -StartupType Disabled }

RUN powershell.exe -command Set-Service -Name lmhosts -StartupType Manual