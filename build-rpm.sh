#!/bin/bash
## ================================================================
## JXWatcher RPM Package Build Script
## ================================================================
## This script builds a RPM (.rpm) package for JXWatcher and places
## the output in the build/ directory.
##
## Required dependencies:
##   sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev rpm
##
## Usage:
##   ./build-rpm.sh [debug|local|local-debug]
##
## WARNING:
##   This script builds an RPM package on a Debian-based system.
##   While functional, it has not been thoroughly tested in this environment.
##   For best results and to avoid potential issues with linked library compatibility,
##   it is strongly recommended to build RPMs on an RPM-based distribution (e.g., Fedora, CentOS, RHEL).
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

echo_start "Starting RPM package build process..."

# Check version.txt
if [ ! -f version.txt ]; then
    echo_error "version.txt not found"
    exit 1
fi

version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo_error "Version not found in version.txt"
    exit 1
fi

# Build flags
ldflags="-w -s"
gcflags="-l"
tags="production,desktop"
cflags="-Os -ffunction-sections -fdata-sections -flto=auto -pipe -pthread"
cldflags="-pthread -Wl,--gc-sections -flto=auto -fwhole-program"

if [[ $1 == "debug" || $1 == "local-debug" ]]; then
    ldflags=""
    gcflags="-l"
    tags="desktop"
    cflags="-pipe -Wall -g -pthread"
    cldflags="-pthread"
    echo_success "Debug mode enabled: building with debug flags"
fi

if [[ $1 == "local" ]]; then
    tags="production,desktop,local"
fi

if [[ $1 == "local-debug" ]]; then
    tags="desktop,local"
fi

# Paths
build_root="build"
rpm_root="${build_root}/rpmbuild"
pkg_root="${build_root}/pkgroot"
bin_path="${pkg_root}/usr/bin"
desktop_path="${pkg_root}/usr/share/applications"
icons_path="${pkg_root}/usr/share/icons/hicolor"

mkdir -p "${bin_path}" "${desktop_path}" \
         "${icons_path}/scalable/apps" \
         "${icons_path}/32x32/apps" \
         "${icons_path}/256x256/apps"

# Build binary
CGO_ENABLED=1 \
CGO_CFLAGS="${cflags}" \
CGO_LDFLAGS="${cldflags}" \
go build -tags="${tags}" -ldflags="${ldflags}" -gcflags="${gcflags}" -o "${pkg_root}/jxwatcher" .

# Copy assets
cp assets/scalable/jxwatcher.svg "${icons_path}/scalable/apps/"
cp assets/32x32/jxwatcher.png "${icons_path}/32x32/apps/"
cp assets/256x256/jxwatcher.png "${icons_path}/256x256/apps/"

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

# Create RPM layout
mkdir -p "${rpm_root}/SPECS" "${rpm_root}/BUILD" "${rpm_root}/RPMS" "${rpm_root}/SOURCES"

# Create tarball source
tar czf "${rpm_root}/SOURCES/jxwatcher-${version}.tar.gz" -C "${pkg_root}" .

# Create .spec file
cat > "${rpm_root}/SPECS/jxwatcher.spec" <<EOF
Name:           jxwatcher
Version:        ${version}
Release:        1
Summary:        Cryptocurrency watcher
License:        MIT
Source0:        jxwatcher-${version}.tar.gz
BuildArch:      x86_64

%description
JXWatcher provides real-time updates and monitoring of various cryptocurrencies.

%prep
%setup -q -c -T
tar -xzf %{SOURCE0}

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/applications
mkdir -p %{buildroot}/usr/share/icons/hicolor/scalable/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/32x32/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/256x256/apps

cp jxwatcher %{buildroot}/usr/bin/jxwatcher
cp usr/share/applications/jxwatcher.desktop %{buildroot}/usr/share/applications/
cp usr/share/icons/hicolor/scalable/apps/jxwatcher.svg %{buildroot}/usr/share/icons/hicolor/scalable/apps/
cp usr/share/icons/hicolor/32x32/apps/jxwatcher.png %{buildroot}/usr/share/icons/hicolor/32x32/apps/
cp usr/share/icons/hicolor/256x256/apps/jxwatcher.png %{buildroot}/usr/share/icons/hicolor/256x256/apps/

%files
/usr/bin/jxwatcher
/usr/share/applications/jxwatcher.desktop
/usr/share/icons/hicolor/scalable/apps/jxwatcher.svg
/usr/share/icons/hicolor/32x32/apps/jxwatcher.png
/usr/share/icons/hicolor/256x256/apps/jxwatcher.png
EOF

# Build RPM
rpmbuild --define "_topdir $(pwd)/${rpm_root}" -bb "${rpm_root}/SPECS/jxwatcher.spec"

if [ $? -ne 0 ]; then
    echo_error "Failed to build the RPM package. Please check for errors above."
    rm -rf "${pkg_root}" "${rpm_root}"
    exit 1
fi

# Move RPM to build/
rpm_file=$(find "${rpm_root}/RPMS" -name "*.rpm" | head -n 1)
mv "$rpm_file" "${build_root}/jxwatcher_${version}_amd64.rpm"

# Clean up temp folders
rm -rf "${pkg_root}" "${rpm_root}"

echo_success "RPM package successfully created at: ${build_root}/jxwatcher_${version}_amd64.rpm"