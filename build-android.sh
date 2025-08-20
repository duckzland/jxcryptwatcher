#!/bin/bash

## ================================================================
## JXWatcher Build Environment Setup Instructions
## ================================================================
##
## Install these requirements package first:
## sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
## sudo apt install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu google-android-ndk-r26c-installer
## go install fyne.io/tools/cmd/fyne@latest
##
## Note:
## - The source of the Android NDK is important, as the compiler path may vary depending on whether it's installed directly from Google.

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

echo_start "Building Android package"

# Check if fyne is installed
if ! command -v fyne &> /dev/null; then
    echo_error "Fyne CLI not found. Please install it using 'go install fyne.io/tools/cmd/fyne@latest'."
    exit 1
fi

# Check if fyne package is available
if ! fyne package -h &> /dev/null; then
    echo_error "Fyne package command not found. Please ensure you have the latest version of Fyne CLI installed."
    exit 1
fi

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo_error "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

# Ubuntu will install android ndk to this folder
android_sdk="/usr/lib/android-ndk/"
if [ ! -d "$android_sdk" ]; then
    echo_error "Android NDK not found at $android_sdk. Please install it."
    exit 1
fi

# Options
name="JXWatcher"
icon="assets/256x256/jxwatcher.png"

# Dynamic variable based on source code
id=$(grep -oP 'const\s+AppID\s*=\s*"\K[^"]+' core/android.go)
if [[ -z "$id" ]]; then
    echo_error "App id not found in core/android.go"
    exit 1
fi

version=$(grep '^version=' version.txt | cut -d'=' -f2)
if [[ -z "$version" ]]; then
    echo_error "Version not found in version.txt"
    exit 1
fi

## Production options 
tags="production,jxandroid"
release="true"

## Note: must have at least -pthread
cflags="-Os -ffunction-sections -fdata-sections -flto=auto -pipe -pthread"
cldflags="-pthread -Wl,--gc-sections -flto=auto -fwhole-program"

## Debugging options, you will need to set -release to false
# tags="jxandroid"
# release="false"

## Note: must have at least -pthread
# cflags="-pipe -Wall -pthread"
# cldflags="-pthread"

## Target os, this will create only for android with arm64 processor
os="android/arm64"

## This will build all possible combination for android, thus big file size
# os="android"

apk_output="build/jxwatcher-$version-android-arm64.apk"

## Generate the AndroidManifest.xml file
cat > AndroidManifest.xml <<EOF
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="io.fyne.jxwatcher">

    <!-- Permissions -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.READ_MEDIA_IMAGES" />
    <uses-permission android:name="android.permission.READ_MEDIA_VIDEO" />
    <uses-permission android:name="android.permission.READ_MEDIA_AUDIO" />
    <uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE"
        android:maxSdkVersion="32" />
    <uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE"
        android:maxSdkVersion="29" />

    <!-- Application block -->
    <application android:label="$name">
        <activity android:name="org.golang.app.GoNativeActivity"
            android:label="$name"
            android:configChanges="orientation|screenSize"
            android:theme="@android:style/Theme.NoTitleBar.Fullscreen">
            
            <meta-data android:name="android.app.lib_name" android:value="fyneapp" />

            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
EOF


CGO_CFLAGS="${cflags}" \
CGO_LDFLAGS="${cldflags}" \
ANDROID_NDK_HOME=$android_sdk \
fyne package -os $os -app-id $id -icon $icon -name $name -app-version $version -tags $tags -release $release

if [ $? -ne 0 ]; then
    echo_error "Failed to package the application."
    rm -f AndroidManifest.xml
    rm -f JXWatcher.apk
    exit 1
fi

mv JXWatcher.apk $apk_output
rm -f AndroidManifest.xml

echo_success "APK package created at: ${apk_output}"