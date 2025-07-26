#!/bin/bash

## ================================================================
## Android Cross-Compilation & APK Build Environment Setup
## ================================================================

## NDK Setup for Compiling Go to Android arm64
## Ensure these packages are installed:
##   sudo apt install gcc-aarch64-linux-gnu \
##                    g++-aarch64-linux-gnu \
##                    google-android-ndk-r26c-installer


###### This script can compile correctly, but I don't understand on how android wont give proper permissions for file operation under android?
###### Anyone can help figure this out?
###### Disabling the script for now..

echo "Script disabled..."
exit


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
app_tags="production, mobile"
# app_flags="-w -s"

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

ANDROID_NDK_HOME=$android_sdk GOFLAGS="$go_flags" fyne package -os android -app-id $app_id -icon $app_icon -name $app_name -app-version $app_version -tags $app_tags -release true

mv JXWatcher.apk build/jxwatcher-$version-android-arm64.apk
rm -rf AndroidManifest.xml