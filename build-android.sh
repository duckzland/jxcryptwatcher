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
if [ ! -d "$android_sdk" ]; then
    echo "Android NDK not found at $android_sdk. Please install it."
    exit 1
fi

# Options
app_id="io.fyne.jxwatcher"
app_name="JXWatcher"
app_version=$(grep '^version=' version.txt | cut -d'=' -f2)
app_icon="assets/256x256/jxwatcher.png"
app_source="$PWD"

## Production options 
app_tags="production,jxandroid"

## Debugging options 
##app_tags="jxmobile"

# Check if fyne is installed
if ! command -v fyne &> /dev/null; then
    echo "Fyne CLI not found. Please install it using 'go install fyne.io/fyne/v2/cmd/fyne@latest'."
    exit 1
fi

# Check if fyne package is available
if ! fyne package -h &> /dev/null; then
    echo "Fyne package command not found. Please ensure you have the latest version of Fyne CLI installed."
    exit 1
fi

# Check if version.txt exists and read the version
if [ ! -f version.txt ]; then
    echo "version.txt not found. Please create a version.txt file with the format 'version=1.0.0'."
    exit 1
fi

version=$(grep '^version=' version.txt | cut -d'=' -f2)

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
    <application android:label="$app_name">
        <activity android:name="org.golang.app.GoNativeActivity"
            android:label="$app_name"
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

ANDROID_NDK_HOME=$android_sdk fyne package -os android -app-id $app_id -icon $app_icon -name $app_name -app-version $app_version -tags $app_tags -release

if [ $? -ne 0 ]; then
    echo "Failed to package the application."
    rm AndroidManifest.xml
    rm JXWatcher.apk
    exit 1
fi

mv JXWatcher.apk build/jxwatcher-$version-android-arm64.apk
rm AndroidManifest.xml