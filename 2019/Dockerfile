ARG BASE_IMAGE_DIGEST

FROM mcr.microsoft.com/windows/servercore:1809$BASE_IMAGE_DIGEST

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

COPY vcredist-2010.x86.exe /vcredist-2010.x86.exe
RUN cmd.exe /s /c "c:\vcredist-2010.x86.exe /install /passive /norestart /wait"
RUN del /F "c:\vcredist-2010.x86.exe"

COPY vcredist-2010.x64.exe /vcredist-2010.x64.exe
RUN cmd.exe /s /c "c:\vcredist-2010.x64.exe /install /passive /norestart /wait"
RUN del /F "c:\vcredist-2010.x64.exe"

COPY vcredist-ucrt.x86.exe /vcredist-ucrt.x86.exe
RUN cmd.exe /s /c "c:\vcredist-ucrt.x86.exe /install /passive /norestart /wait"
RUN del /F "c:\vcredist-ucrt.x86.exe"

COPY vcredist-ucrt.x64.exe /vcredist-ucrt.x64.exe
RUN cmd.exe /s /c "c:\vcredist-ucrt.x64.exe /install /passive /norestart /wait"
RUN del /F "c:\vcredist-ucrt.x64.exe"

# Install .NET Framework 4.8 (1809 ships with 4.7)
COPY dotnet-48-installer.exe /dotnet-48-installer.exe
RUN cmd.exe /s /c "c:\dotnet-48-installer.exe /q"
RUN del /F "c:\dotnet-48-installer.exe"

RUN powershell.exe Remove-Item -Force -Recurse ${Env:TEMP}\*

RUN powershell.exe -command "remove-windowsfeature -name 'windows-defender'"

# disable common unneeded services
RUN powershell.exe -command \
  $svs=('AppHostSvc', 'MSDTC', 'TermService', 'WAS', 'dhcp', 'diagtrack', 'w3svc', 'winrm', 'RemoteRegistry'); \
  foreach ($name in $svs) { Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\$name" -Name Start -Value 4 }

# disable 2019 specific additions/changes
RUN powershell.exe -command \
  $svs=('Sense', 'SCardSvr', 'UsoSvc', 'SysMain', 'SgrmBroker', 'AppReadiness'); \
  foreach ($name in $svs) { Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\$name" -Name Start -Value 4 }

RUN powershell.exe -command Set-Service -Name lmhosts -StartupType Manual

# enable automatic start of DNS cache to support FQDNs for net use
RUN powershell.exe -command "Set-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\dnscache' -Name Start -Value 2"

# 10s of graceful shutdown time
RUN reg add hklm\system\currentcontrolset\services\cexecsvc /v ProcessShutdownTimeoutSeconds /t REG_DWORD /d 10 && \
    reg add hklm\system\currentcontrolset\control /v WaitToKillServiceTimeout /t REG_SZ /d 10000 /f

# enable ODBC registry
RUN powershell.exe -command \ 
  $acl = Get-Acl HKLM:\SOFTWARE\ODBC; \
  $rule = New-Object System.Security.AccessControl.RegistryAccessRule('vcap', 'WriteKey', 'ContainerInherit', 'None', 'Allow') ; \
  $acl.SetAccessRule($rule); \
  Set-Acl -AclObject $acl -Path HKLM:\SOFTWARE\ODBC;
