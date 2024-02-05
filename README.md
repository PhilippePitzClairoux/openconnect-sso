# Go Openconnect SSO
This has been tested on fedora only.
You need openconnect and chromium installed (and golang to debug).

In order to build this project, you only need docker.
## How to build
```bash
./build.sh
```

You can set env variables with the name GOOS and GOARCH in order to change
the OS and cpu architecture of the build.
```bash
GOOS=darwin GOARCH=arm ./build.sh
```

## Usage
```
./go-openconnect-sso --server vpn.host.com
```