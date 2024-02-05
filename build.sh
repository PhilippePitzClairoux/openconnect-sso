#!/usr/bin/env bash

# Tag dockerfile
docker build -t openconnect-sso-builder .

# Run builder to create executable
docker run --name builder openconnect-sso-builder

# Copy output directory to current directory
docker cp builder:/opt/out ./out