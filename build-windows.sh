#!/bin/bash

## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================
##
## Install these requirements package first:
## 
## sudo apt install gcc-mingw-w64 wixl uuid-runtime
## go install github.com/tc-hib/go-winres@latest
##
## Note:
## - uuid-runtime is optional, only if you wish to regenerate new uuid

set -e

# Check if go-winres is installed
if ! command -v go-winres &> /dev/null; then
    echo "Command go-winres not found, install with 'go install github.com/tc-hib/go-winres@latest'"
    exit 1
fi

# Check if wixl is installed
# if ! command -v wixl &> /dev/null; then
#     echo "Command wixl not found, install with 'sudo apt install wixl'"
#     exit 1
# fi

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

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
manufacturer="duckzland"
bin_name="jxwatcher.exe"
manufacturer="duckzland"
build_dir="build"
msi_output="${build_dir}/${app_name}-${version}.msi"
wxs_file="${build_dir}/installer.wxs"

# Only no need to regenerate this
# upgrade_guid=$(uuidgen)
upgrade_guid="11442397-0932-48db-98f3-eeb7704e669a"
if [[ -z "$upgrade_guid" ]]; then
    echo "Command uuidgen not found, install with 'sudo apt install uuid-runtime'"
    exit 1
fi

# Change component GUIDs only when necessary: A new component GUID is required if the key path of 
# the component changes (e.g., a file is renamed or moved to a different location) 
# or if the component's content fundamentally changes such that it's no longer the same logical entity.
# component_guid=$(uuidgen)
component_guid="900c3a3d-77b4-468c-8a2d-8a6eb9dd838e"
if [[ -z "$component_guid" ]]; then
    echo "Command uuidgen not found, install with 'sudo apt install uuid-runtime'"
    exit 1
fi

# Whether you bump from 1.0.0 → 1.0.1 or 1.0.0 → 2.0.0, if your MSI should uninstall the old version and 
# install the new one, you need a new Product Id every time.
# Generate one manually or use uuidgen if you have one installed
product_guid=$(uuidgen)
# product_guid="0d50ab79-1abd-4277-8b96-d92fffa1afa6"
if [[ -z "$product_guid" ]]; then
    echo "Command uuidgen not found, install with 'sudo apt install uuid-runtime'"
    exit 1
fi

# The Guid assigned to the shortcut's <Component> (e.g. AppShortcut) should remain the same across builds, 
# unless something structurally changes in the shortcut component.
# shortcut_guid=$(uuidgen)
shortcut_guid="3e0298de-9989-49c6-9285-23c7e55399b0"
if [[ -z "$shortcut_guid" ]]; then
    echo "Command uuidgen not found, install with 'sudo apt install uuid-runtime'"
    exit 1
fi




# Build Go binary
GOOS=windows \
GOARCH=amd64  \
CGO_ENABLED=1 \
CGO_CFLAGS="-pthread" \
CGO_LDFLAGS="-pthread" \
CC=/usr/bin/x86_64-w64-mingw32-gcc \
go build -tags="production,desktop" -ldflags "-w -s -H=windowsgui" -o "${build_dir}/${bin_name}" .

echo "Windows binary generated: ${build_dir}/${bin_name}"

# Copy Windows resource file
rsrc_file="assets/windows/rsrc_windows_amd64.syso"

# Update the syso file
cd assets/windows/
go-winres simply --icon jxwatcher.ico --product-version $version --file-version $version --product-name $app_name
cd ../../

cp "$rsrc_file" rsrc_windows_amd64.syso
rm assets/windows/rsrc_windows*

# Remove copied resource file
rm -f rsrc_windows_amd64.syso

# Create WXS (installer source) file
cat > "$wxs_file" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">

  <Product Id="$product_guid" Name="$app_name" Language="1033" Version="$version"
           Manufacturer="$manufacturer" UpgradeCode="$upgrade_guid">

    <Package InstallerVersion="500" Compressed="yes" InstallScope="perMachine" />
    <Property Id="ARPPRODUCTICON" Value="jxwatcher.ico" />

    <!-- Define upgrade behavior -->
    <Upgrade Id="$upgrade_guid">
      <UpgradeVersion Minimum="0.0.0" Maximum="$version" OnlyDetect="no" Property="OLD_VERSION_FOUND" />
      <UpgradeVersion Minimum="$version" IncludeMinimum="yes" OnlyDetect="yes" Property="NEWER_VERSION_FOUND" />
    </Upgrade>

    <InstallExecuteSequence>
      <RemoveExistingProducts After="InstallInitialize" />
    </InstallExecuteSequence>

    <Media Id="1" Cabinet="media1.cab" EmbedCab="yes" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="$app_name">
          <Component Id="MainBinary" Guid="$component_guid">
            <File Id="AppBinary" Source="${bin_name}" Name="${bin_name}" KeyPath="yes" />
          </Component>
        </Directory>
      </Directory>

      <!-- Start Menu Directory -->
      <Directory Id="ProgramMenuFolder">
        <Directory Id="AppProgramMenuDir" Name="$app_name">
          <Component Id="AppShortcut" Guid="$shortcut_guid">
            <Shortcut Id="startMenuShortcut"
                      Name="$app_name"
                      Description="$app_description"
                      Target="[INSTALLFOLDER]\\${bin_name}"
                      WorkingDirectory="INSTALLFOLDER"
                      Icon="jxwatcher.ico"
                      IconIndex="0" />
            <RemoveFolder Id="AppProgramMenuDir" On="uninstall"/>
            <RegistryValue Root="HKCU" Key="Software\\${manufacturer}\\${app_name}" Name="installed"
                           Type="integer" Value="1" KeyPath="yes"/>
          </Component>
        </Directory>
      </Directory>
    </Directory>

    <Feature Id="DefaultFeature" Level="1">
      <ComponentRef Id="MainBinary" />
      <ComponentRef Id="AppShortcut" />
    </Feature>

    <Icon Id="jxwatcher.ico" SourceFile="assets/windows/jxwatcher.ico" />

  </Product>
</Wix>
EOF

# Build MSI
echo "Building MSI..."
wixl -o "${msi_output}" "${wxs_file}"

echo "MSI package created: ${msi_output}"

# Clean up
rm -f "$wxs_file"