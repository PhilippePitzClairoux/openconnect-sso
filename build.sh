#!/usr/bin/env bash

# Remove old instance
echo -e "Removing old builds (if present)\n"
docker rm builder
docker rmi openconnect-sso-builder
rm -rf out

# Tag dockerfile
echo -e "Build and tag container\n"
docker build -t openconnect-sso-builder .

OVERRIDE_GO_VARIABLES=""

if [ -n "${GOOS}" ]; then
  OVERRIDE_GO_VARIABLES+=" -e \"GOOS=${GOOS}\""
fi

if [ -n "${GOARCH}" ]; then
  OVERRIDE_GO_VARIABLES+=" -e \"GOARCH=${GOARCH}\""
fi

# Run builder to create executable
echo -e "Run container and assign it the name \"builder\"\n"
sh -c "docker run  ${OVERRIDE_GO_VARIABLES} --name builder openconnect-sso-builder"

# Copy output directory to current directory
echo -e "Copy executable from container to ./out directory\n"
docker cp builder:/opt/out ./out

echo "Copying ./out/openconnect-sso to /usr/bin/openconnect-sso"
sudo cp ./out/openconnect-sso /usr/bin/openconnect-sso