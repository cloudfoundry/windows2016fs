Integrated Windows Authentication Sample App
====

.NET api app for testing Integrated Windows Auth on CF


Deploy
=======

Run the following command to deploy:

```sh
cf push windows-auth -s windows -b hwc_buildpack
```

Requirements
=======
This sample app requires at least 512mb of memory to run on CloudFoundry.


Endpoints
=======
1. `GET /` Howdy!
1. `GET /auth` Returns the name and authentication type of the authenticated user


Authenticating
======
To authenticate via the browser, simply visit the `/auth` endpoint and you'll be prompted for credentials.  For example, if your domain is `MSFT` and your credentials are `bgates:m0n3ybag$`, you would enter `MSFT\bgates` as your username and `m0n3ybag$` as your password.

To use `curl`, you would do something like
```
curl --ntlm --user 'MSFT\bgates:m0n3ybag$' https://auth.apps.<cf-instance>/auth
```


Building
=============

#### Requirements:

* Microsoft Windows OS

* [Msbuild.exe](https://docs.microsoft.com/en-us/visualstudio/msbuild/msbuild)

#### Build

##### Method 1

* Make sure you have `msbuild.exe` on your `$PATH`. (If you're using a VM created using a [BOSH Stemcell for Windows](https://bosh.io/stemcells), it is available at `$env:WINDIR\Microsoft.NET\Framework64\v*\MSBuild.exe`)

* Make your code changes

* Run `./make.bat`

* This will build the app and place it in the newly created `bin/` directory.

##### Method 2

You can also use the dotnet-framework docker image which has `msbuild` preinstalled to easily build the app.

* Make your code changes

* Pull the docker image: `docker pull mcr.microsoft.com/dotnet/framework/sdk`

* Run a container with this directory mounted: `docker run --rm -it -v <path/to/WindowsAuth>:C:\windows-auth mcr.microsoft.com/dotnet/framework/sdk powershell`

* Inside the container: `cd windows-auth`, and `./make.bat`. This should build the app.

* Exit the container using `exit`.