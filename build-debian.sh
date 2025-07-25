#!/bin/bash

##
## This is the minimal build script that will generate debian package at /build folder
##
## requires the following dependencies:
## - dpkg-deb
##
## sudo apt install dpkg-dev
## 


echo "Generating Debian Package"

mkdir -p build/jxwatcher_1.0.0/DEBIAN
mkdir -p build/jxwatcher_1.0.0/usr/bin
mkdir -p build/jxwatcher_1.0.0/usr/share/applications
mkdir -p build/jxwatcher_1.0.0/usr/share/icons/hicolor/scalable/apps
mkdir -p build/jxwatcher_1.0.0/usr/share/icons/hicolor/32x32/apps

go build -tags production -ldflags "-w -s" -gcflags="-l" -o build/jxwatcher_1.0.0/usr/bin/jxwatcher .

# Create control file
cat <<EOF > build/jxwatcher_1.0.0/DEBIAN/control
Package: JXWatcher
Version: 1.0.0
Section: base
Priority: optional
Architecture: amd64
Maintainer: JXWatcher <nobody@example.com>
Description: JXWatcher is a cryptocurrency watcher application that provides real-time updates and monitoring of various cryptocurrencies.
EOF

# Create Desktop file
cat <<EOF > build/jxwatcher_1.0.0/usr/share/applications/jxwatcher.desktop
[Desktop Entry]
Name=JXWatcher
Exec=/usr/bin/jxwatcher
Icon=jxwatcher
Type=Application
Categories=Utility;
Terminal=false
EOF

# Copy the assets
cp assets/scalable/jxwatcher.svg build/jxwatcher_1.0.0/usr/share/icons/hicolor/scalable/apps/
cp assets/32x32/jxwatcher.png build/jxwatcher_1.0.0/usr/share/icons/hicolor/32x32/apps/


# Build the package
dpkg-deb --build build/jxwatcher_1.0.0 build/jxwatcher_1.0.0.deb

if [ $? -eq 0 ]; then
    echo "Debian package created successfully: build/jxwatcher_1.0.0.deb"
else
    echo "Failed to create Debian package."
fi

# Clean up the build directory
rm -rf build/jxwatcher_1.0.0/

