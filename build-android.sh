#!/bin/bash

## ================================================================
## Android Cross-Compilation & APK Build Environment Setup
## ================================================================

## NDK Setup for Compiling Go to Android arm64
## Ensure these packages are installed:
##   sudo apt install gcc-aarch64-linux-gnu \
##                    g++-aarch64-linux-gnu \
##                    google-android-ndk-r26c-installer

set -e

# Ubuntu ndk will be installed in this folder
android_sdk="/usr/lib/android-ndk/"

# Without this flags the binary size will be large, still havent found the clue on how to inject this flag
# when using fyne package
# go_flags="-ldflags='-w -s'"

# Options
app_id="io.github.duckzland.jxcryptwatcher"
app_name="JXWatcher"
app_version=$(grep '^version=' version.txt | cut -d'=' -f2)
app_icon="assets/256x256/jxwatcher.png"
app_source="$PWD"
app_tags="production"
# app_flags="-w -s"

version=$(grep '^version=' version.txt | cut -d'=' -f2)

ANDROID_NDK_HOME=$android_sdk GOFLAGS="$go_flags" fyne package -os android -app-id $app_id -icon $app_icon -name $app_name -app-version $app_version -tags $app_tags -release true

mv JXWatcher.apk build/jxwatcher-$version-android-arm64.apk
