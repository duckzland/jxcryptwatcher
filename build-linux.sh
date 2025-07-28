#!/bin/bash

## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================
##
## This is the minimal build script that will generate linux binary at /build folder
##
## flags -ldflags "-w -s" -gcflags="-l" is for creating smallest possible file
## 

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

echo "Generating Linux binary"

# Load version info
version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo "Error: Version not found in version.txt"
    exit 1
fi

target_output="build/jxwatcher-$version-linux-amd64"

go build -tags="production,desktop" -ldflags "-w -s" -gcflags="-l" -o $target_output .

echo "Linux binary created: ${target_output}"