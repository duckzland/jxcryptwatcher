#!/bin/bash

## ================================================================
## JXWatcher Debian Package Build Script Overview
## ================================================================

## Description:
## This script builds a Debian (.deb) package and places the output
## in the /build directory.

## Dependencies:
## - dpkg-deb             : Used to build the Debian package
## - Go                   : Must be installed and available in PATH
## - Assets directory     : Required; must be present
## - version.txt file     : Required; must contain version info

set -euo pipefail

echo "Generating Debian Package..."

# Extract version
if [[ ! -f version.txt ]]; then
    echo "version.txt not found."
    exit 1
fi

version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo "Version not found in version.txt"
    exit 1
fi

# Define paths
pkg_dir="build/jxwatcher-${version}"
bin_path="${pkg_dir}/usr/bin"
icons_path="${pkg_dir}/usr/share/icons/hicolor"
desktop_path="${pkg_dir}/usr/share/applications"
deb_output="build/jxwatcher_${version}_amd64.deb"

# Create necessary directories
mkdir -p "${pkg_dir}/DEBIAN" \
         "${bin_path}" \
         "${desktop_path}" \
         "${icons_path}/scalable/apps" \
         "${icons_path}/32x32/apps"

# Build the Go binary
echo "Building Go binary..."
go build -tags production,desktop -ldflags "-w -s" -gcflags="-l" -o "${bin_path}/jxwatcher" .

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
echo "Copying assets..."
cp assets/scalable/jxwatcher.svg "${icons_path}/scalable/apps/"
cp assets/32x32/jxwatcher.png "${icons_path}/32x32/apps/"

# Build the Debian package
echo "Building Debian package..."
dpkg-deb --build "${pkg_dir}" "${deb_output}"

echo "Package created successfully: ${deb_output}"

# Clean up
echo "Cleaning up..."
rm -rf "${pkg_dir}"