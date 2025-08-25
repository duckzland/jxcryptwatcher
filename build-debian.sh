#!/bin/bash

## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================
##
## This script builds a Debian (.deb) package and places the output
## in the /build directory.
##
## Install requirements:
## sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
## sudo apt install dpkg
##
## Note: 
## - Ubuntu should already have dpkg package installed by default

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

echo_start "Building Debian package"

# Check if go-winres is installed
if ! command -v dpkg-deb &> /dev/null; then
    echo_error "Command dpkg-deb not found, install with 'apt install dpkg'"
    exit 1
fi

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo_error "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo_error "Version not found in version.txt"
    exit 1
fi

# Define paths
pkg_dir="build/jxwatcher-${version}"
bin_path="${pkg_dir}/usr/bin"
icons_path="${pkg_dir}/usr/share/icons/hicolor"
desktop_path="${pkg_dir}/usr/share/applications"
deb_output="build/jxwatcher_${version}_amd64.deb"

# Production compiling flags
ldflags="-w -s"
gcflags="-l"
tags="production,desktop"
cflags="-Os -ffunction-sections -fdata-sections -flto=auto -pipe -pthread"
cldflags="-pthread -Wl,--gc-sections -flto=auto -fwhole-program"

# Debug compiling flags
if [[ $1 == "debug" ]]; then
    ldflags=""
    gcflags="-l"
    tags="desktop"
    cflags="-pipe -Wall -g -pthread"
    cldflags="-pthread"
    echo_success "Generating binary for debugging"
fi

# Create necessary directories
mkdir -p "${pkg_dir}/DEBIAN" \
         "${bin_path}" \
         "${desktop_path}" \
         "${icons_path}/scalable/apps" \
         "${icons_path}/32x32/apps"

# Build the Go binary
CGO_ENABLED=1 \
CGO_CFLAGS="${cflags}" \
CGO_LDFLAGS="${cldflags}" \
go build -tags="${tags}" -ldflags "${ldflags}" -gcflags="${gcflags}" -o "${bin_path}/jxwatcher" .

# Create control file
cat > "${pkg_dir}/DEBIAN/control" <<EOF
Package: jxwatcher
Version: ${version}
Section: base
Priority: optional
Architecture: amd64
Maintainer: JXWatcher <nobody@example.com>
Description: JXWatcher is a cryptocurrency watcher application that provides real-time updates and monitoring of various cryptocurrencies.
EOF

# Create desktop entry
cat > "${desktop_path}/jxwatcher.desktop" <<EOF
[Desktop Entry]
Name=JXWatcher
Exec=/usr/bin/jxwatcher
Icon=jxwatcher
Type=Application
Categories=Utility;
Terminal=false
EOF

# Copy assets
cp assets/scalable/jxwatcher.svg "${icons_path}/scalable/apps/"
cp assets/32x32/jxwatcher.png "${icons_path}/32x32/apps/"

# Build the Debian package
dpkg-deb --build "${pkg_dir}" "${deb_output}" &> /dev/null

if [ $? -ne 0 ]; then
    echo_error "Failed to package the application."
    rm -rf "${pkg_dir}"
    exit 1
fi

# Clean up
rm -rf "${pkg_dir}"

echo_success "Debian package created at: ${deb_output}"