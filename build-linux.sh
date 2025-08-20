#!/bin/bash

## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================
##
## Install requirements:
## sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
##
## This is the minimal build script that will generate linux binary at /build folder
##
## flags -ldflags "-w -s" -gcflags="-l" is the minimum flags for small binary output
## 

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

echo -e "\033[1mGenerating Linux binary\033[0m"

# Load version info
version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo "Error: Version not found in version.txt"
    exit 1
fi

target_output="build/jxwatcher-$version-linux-amd64"

# Production compiling flags
ldflags="-w -s"
gcflags="-l"
tags="production,desktop"
cflags="-Os -ffunction-sections -fdata-sections -flto=auto -pipe -pthread"
cldflags="-pthread -Wl,--gc-sections -flto=auto -fwhole-program"

# Debug compiling flags
# ldflags=""
# gcflags="-l"
# tags="desktop"
# cflags="-pipe -Wall -g -pthread"
# cldflags="-pthread"

CGO_ENABLED=1 \
CGO_CFLAGS="${cflags}" \
CGO_LDFLAGS="${cldflags}" \
go build -tags="${tags}" -ldflags "${ldflags}" -gcflags="${gcflags}" -o $target_output .

echo "Linux binary created: ${target_output}"