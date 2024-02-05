# Go Openconnect SSO
This has been tested on fedora only.
You need openconnect and chromium installed (and golang to debug).

In order to build this project, you only need docker.
## How to install
if you already have golang installed on your system, you can simply run the following
command in order to install this application.

```bash
go install github.com/PhilippePitzClairoux/openconnect-sso@latest
```

[for more information about `go install`, click me!](https://go.dev/ref/mod#go-install)

## Usage
Simple example :
```bash
openconnect-sso --server vpn.host.com
```

Auto-fill username and/or password
```bash
openconnect-sso --server vpn.host.com --username myuser@email.com --password oopsThisMightNotBeTheBestIdeaEver
```

## How to build
```bash
./build.sh
```

You can set env variables with the name GOOS and GOARCH in order to change
the OS and cpu architecture of the build.
```bash
GOOS=darwin GOARCH=arm ./build.sh
```

To build from scratch you can use the following command
```bash
go build ./...
```