:: msbuild must be in path
SET PATH=%PATH%;%WINDIR%\Microsoft.NET\Framework64\v4.0.30319
where msbuild
if errorLevel 1 ( echo "msbuild was not found on PATH" && exit /b 1 )

rmdir /S /Q packages
.nuget\nuget restore || exit /b 1
MSBuild WindowsAuth.sln /t:Rebuild /p:Configuration=Release || exit /b 1