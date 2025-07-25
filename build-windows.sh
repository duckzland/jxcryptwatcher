#!/bin/bash

## 
## You must install mingw for compiling go to windows binary
## sudo apt install gcc-mingw-w64
##

## To regenerate the syso file, you need to run the following command:
## go-winres simply --icon jxwatcher.ico --product-version 1.0.0 --file-version 1.0.0 --product-name "JXWatcher"
##
## You can install go-winres with the following command:
## go install github.com/tc-hib/go-winres@latest

## To regenerate the .ico file, you can copy the following function into your .bashrc or bash_profile
## and then run the command `svgtoico jxwatcher scalable/jxwatcher.svg windows/jxwatcher.ico`
## This function uses Inkscape to convert SVG to PNG and ImageMagick to convert PNGs to ICO.
##
## Make sure you have Inkscape and ImageMagick installed:
## sudo apt install inkscape imagemagick    
##
# svgtoico(){
#     # $1: Base name for the icon (e.g., "my_icon")
#     # $2: Path to the source SVG file (e.g., "path/to/my_icon.svg")
#     # $3: Desired path for the output ICO file (e.g., "path/to/output.ico")

#     # Create temporary PNG files for different sizes
#     inkscape -w 16 -h 16 -o "$1-16.png" "$2"
#     inkscape -w 32 -h 32 -o "$1-32.png" "$2"
#     inkscape -w 48 -h 48 -o "$1-48.png" "$2"
#     inkscape -w 256 -h 256 -o "$1-256.png" "$2"

#     # Combine PNGs into an ICO file
#     convert "$1-16.png" "$1-32.png" "$1-48.png" "$1-256.png" "$1.ico"

#     # Clean up temporary PNG files
#     rm "$1-16.png" "$1-32.png" "$1-48.png" "$1-256.png"

#     # Move the generated ICO to the desired location
#     mv "$1.ico" "$3"
# }
##
## To build the JXWatcher.msi you need these dependencies:
## sudo apt install wixl msitools uuid-runtime


## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================

## 1. Cross-Compiling Go to Windows Binary
##    - Requirement: MinGW-w64
##    - Install with: sudo apt install gcc-mingw-w64

## 2. Regenerating the .syso File
##    - Tool: go-winres
##    - Command:
##      go-winres simply \
##         --icon jxwatcher.ico \
##         --product-version 1.0.0 \
##         --file-version 1.0.0 \
##         --product-name "JXWatcher"
##    - Install go-winres:
##      go install github.com/tc-hib/go-winres@latest

## 3. Regenerating the .ico File from SVG
##    - Required tools: Inkscape, ImageMagick
##    - Install with:
##      sudo apt install inkscape imagemagick
##
##    - Suggested shell function (add to .bashrc or .bash_profile):
##
##      svgtoico() {
##          # $1: Base name (e.g., "my_icon")
##          # $2: Source SVG path
##          # $3: Destination ICO path
##
##          inkscape -w 16 -h 16 -o "$1-16.png" "$2"
##          inkscape -w 32 -h 32 -o "$1-32.png" "$2"
##          inkscape -w 48 -h 48 -o "$1-48.png" "$2"
##          inkscape -w 256 -h 256 -o "$1-256.png" "$2"
##
##          convert "$1-16.png" "$1-32.png" "$1-48.png" "$1-256.png" "$1.ico"
##
##          rm "$1-16.png" "$1-32.png" "$1-48.png" "$1-256.png"
##          mv "$1.ico" "$3"
##      }

## 4. Building JXWatcher.msi Installer
##    - Dependencies: wixl, msitools, uuid-runtime
##    - Install with:
##      sudo apt install wixl msitools uuid-runtime

set -euo pipefail

echo "Building Windows binary..."

# Load version info
version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
if [[ -z "$version" ]]; then
    echo "Error: Version not found in version.txt"
    exit 1
fi

# App metadata
app_name="JXWatcher"
package_name="io.github.duckzland.jxcryptwatcher"
bin_name="jxwatcher.exe"
manufacturer="duckzland"
build_dir="build"
msi_output="${build_dir}/${app_name}.msi"
wxs_file="${build_dir}/installer.wxs"

# Generate GUIDs
upgrade_guid=$(uuidgen)
component_guid=$(uuidgen)

# Copy Windows resource file
rsrc_file="assets/windows/rsrc_windows_amd64.syso"
if [[ ! -f "$rsrc_file" ]]; then
    echo "Error: Resource file not found: $rsrc_file"
    exit 1
fi

# Update the syso file
cd assets/windows/
go-winres simply --icon jxwatcher.ico --product-version $version --file-version $version --product-name $app_name
cd ../../

cp "$rsrc_file" rsrc_windows_amd64.syso


# Build Go binary
GOOS=windows \
GOARCH=amd64  \
CGO_ENABLED=1 \
CGO_CFLAGS="-pthread" \
CGO_LDFLAGS="-pthread" \
CC=/usr/bin/x86_64-w64-mingw32-gcc \
go build -tags production -ldflags "-w -s -H=windowsgui" -o "${build_dir}/${bin_name}" .

echo "Windows binary generated: ${build_dir}/${bin_name}"

# Remove copied resource file
rm -f rsrc_windows_amd64.syso

# Create WXS (installer source) file
cat > "$wxs_file" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*" Name="$app_name" Language="1033" Version="$version"
           Manufacturer="$manufacturer" UpgradeCode="$upgrade_guid">
    <Package InstallerVersion="200" Compressed="yes" InstallScope="perMachine" />

    <Media Id="1" Cabinet="media1.cab" EmbedCab="yes" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="$app_name">
          <Component Id="MainBinary" Guid="$component_guid">
            <File Id="AppBinary" Source="${bin_name}" Name="${bin_name}" KeyPath="yes" />
          </Component>
        </Directory>
      </Directory>
    </Directory>

    <Feature Id="DefaultFeature" Level="1">
      <ComponentRef Id="MainBinary" />
    </Feature>
  </Product>
</Wix>
EOF

echo "Installer source generated: ${wxs_file}"
echo "Upgrade GUID: ${upgrade_guid}"
echo "Component GUID: ${component_guid}"

# Verify wixl availability
if ! command -v wixl &> /dev/null; then
    echo "Error: 'wixl' not installed. Install via: sudo apt install msitools"
    exit 1
fi

# Build MSI
echo "Building MSI..."
wixl -o "${msi_output}" "${wxs_file}"

echo "MSI package created: ${msi_output}"

# Clean up
rm -f "$wxs_file"