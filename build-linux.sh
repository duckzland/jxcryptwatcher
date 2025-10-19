#!/bin/bash

## ================================================================
## JXWatcher Linux Build Script
## ================================================================
##
## Required dependencies:
##   sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
##
## This script builds a minimal Linux binary in the /build directory.
##
## For smaller binaries, production flags are used:
##   -ldflags="-w -s" -gcflags="-l"
##
## For debugging, run: ./build-linux.sh debug
##
## ================================================================

set -e

echo_error() {
  echo -e "\033[0;31m- $1\033[0m"
}

echo_success() {
  echo -e "\033[0;32m- $1\033[0m"
}

echo_start() {
  echo -e "\033[1m$1\033[0m"
}

echo_start "Starting Linux build process..."

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo_error "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

# Load version info
version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo_error "Error: Version not found in version.txt"
    exit 1
fi

target_output="build/jxwatcher-$version-linux-amd64"

# Production compiling flags
ldflags="-w -s"
gcflags="-l"
tags="production,desktop,no_emoji,no_animations"
cflags="-Os -ffunction-sections -fdata-sections -flto=auto -pipe -pthread"
cldflags="-pthread -Wl,--gc-sections -flto=auto -fwhole-program"

# Debug compiling flags
if [[ $1 == "debug" || $1 == "local-debug" ]]; then
  ldflags=""
  gcflags="-l"
  tags="desktop,no_emoji,no_animations"
  cflags="-pipe -Wall -g -pthread"
  cldflags="-pthread"

  echo_start "Debug mode enabled: building with debug flags"
fi

if [[ $1 == "local" ]]; then
    tags="production,desktop,local,no_emoji,no_animations"
fi

if [[ $1 == "local-debug" ]]; then
    tags="desktop,local,no_emoji,no_animations"
fi

CGO_ENABLED=1 \
CGO_CFLAGS="${cflags}" \
CGO_LDFLAGS="${cldflags}" \
go build -tags="${tags}" -ldflags "${ldflags}" -gcflags="${gcflags}" -o $target_output .

if [ $? -ne 0 ]; then
    echo_error "Linux binary creation failed. Please check the build output above for details."
    exit 1
fi

echo_success "Linux binary successfully created at: ${target_output}"