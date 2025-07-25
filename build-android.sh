#!/bin/bash

## ================================================================
## Android Cross-Compilation & APK Build Environment Setup
## ================================================================

## NDK Setup for Compiling Go to Android arm64
## Ensure these packages are installed:
##   sudo apt install gcc-aarch64-linux-gnu \
##                    g++-aarch64-linux-gnu \
##                    google-android-ndk-r26c-installer

## Validate required paths:
##   - Sysroot must exist at:
##       /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sys_root
##
##   - Target lib directory for API level 26 must be available:
##       /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sys_root/usr/lib/aarch64-linux-android/26
##
##   - Clang compiler should be found at:
##       /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
##
##   - Match '--target' flag with lib directory version:
##       --target=aarch64-linux-android26

## APK Build Toolchain
## Install essential packaging/signing tools:
##   sudo apt install aapt apksigner zipalign android-sdk \
##                    google-android-build-tools-30.0.3-installer \
##                    google-android-cmdline-tools-11.0-installer

## Fetch SDK components manually:
##   mkdir -p ~/Android/Sdk
##   sdkmanager --sdk_root=$HOME/Android/Sdk "build-tools;30.0.3"
##   sdkmanager --sdk_root=$HOME/Android/Sdk "platforms;android-30"

## Notes:
## - build-tools are required (includes aapt, zipalign, apksigner)
## - Default user path for SDK will be:
##     $HOME/Android/Sdk

echo "Eventhough the compilation and apk build completed, seems it is producing invalid apk. so not supported yet!"
exit 1
set -e

echo "Generating android arm64 binary..."

version=$(grep '^version=' version.txt | cut -d'=' -f2)

ndk_root="/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64"
sys_root="$ndk_root/sysroot"
lib_dir="$sys_root/usr/lib/aarch64-linux-android/26"
clang="$ndk_root/bin/clang"

[[ -d "$sys_root" ]] || { echo "Missing sysroot: $sys_root"; exit 1; }
[[ -d "$lib_dir" ]] || { echo "Missing lib dir: $lib_dir"; exit 1; }
[[ -x "$clang" ]] || { echo "Clang not executable: $clang"; exit 1; }

## The literal string used for building
#GOOS=android GOARCH=arm64 CGO_ENABLED=1 CGO_CFLAGS="--sysroot=/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot -pthread" CGO_LDFLAGS="-L/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android/26 -pthread"  CC="/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang --target=aarch64-linux-android26"  go build -tags production -ldflags="-w -s -linkmode=external -extldflags="-fuse-ld=lld"" -o build/jxwatcher-android-arm64

GOOS=android \
GOARCH=arm64 \
CGO_ENABLED=1 \
CGO_CFLAGS="--sysroot=${sys_root} -pthread" \
CGO_LDFLAGS="-L${lib_dir} -pthread" \
CC="${clang} --target=aarch64-linux-android26" \
go build -tags production \
-ldflags="-w -s -linkmode=external -extldflags="-fuse-ld=lld"" \
-o build/jxwatcher-android-arm64

##
## Build the APK
##
echo "Building APK..."

app_name="JXWatcher"
package_name="io.github.duckzland.jxcryptwatcher"
arch="arm64-v8a"
bin_name="jxwatcher-android-arm64"
keystore="$HOME/.android/debug.keystore"
keyalias="androiddebugkey"
keypass="android"
keystorepass="android"
android_home="$HOME/Android/Sdk"

rm -rf build/apkbuild
mkdir -p build/apkbuild/{lib/$arch,assets,res}

cp "build/$bin_name" "build/apkbuild/lib/$arch/lib$bin_name.so"

# Generate AndroidManifest.xml
cat > build/apkbuild/AndroidManifest.xml <<EOF
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="$package_name"
    android:versionCode="${version//.}"
    android:versionName="$version">
    <application android:label="$app_name">
        <activity android:name=".MainActivity">
            <intent-filter>
                <action android:name="android.intent.action.MAIN"/>
                <category android:name="android.intent.category.LAUNCHER"/>
            </intent-filter>
        </activity>
    </application>
</manifest>
EOF

# Create debug keystore if missing
if [[ ! -f "$keystore" ]]; then
  keytool -genkeypair \
    -alias "$keyalias" \
    -keyalg RSA \
    -keysize 2048 \
    -validity 10000 \
    -keystore "$keystore" \
    -storepass "$keystorepass" \
    -keypass "$keypass" \
    -dname "CN=Android Debug,O=Android,C=US"
fi

# Package, align, and sign APK
aapt package -f \
  -M build/apkbuild/AndroidManifest.xml \
  -F build/apkbuild/$app_name.unaligned.apk \
  -I "$android_home/platforms/android-30/android.jar" \
  -S build/apkbuild/res \
  -A build/apkbuild/assets \
  -m

aapt add build/apkbuild/$app_name.unaligned.apk build/apkbuild/lib/$arch/lib$bin_name.so

zipalign -f 4 \
  build/apkbuild/$app_name.unaligned.apk \
  build/apkbuild/$app_name.aligned.apk

apksigner sign \
  --ks "$keystore" \
  --ks-key-alias "$keyalias" \
  --ks-pass pass:"$keystorepass" \
  --key-pass pass:"$keypass" \
  --out "build/apkbuild/$app_name.apk" \
  build/apkbuild/$app_name.aligned.apk

# Move the file
mv build/apkbuild/$app_name.apk build/
mv build/$bin_name build/$bin_name-$version

echo "APK built: build/apkbuild/$app_name.apk"

# Clean up
rm -rf build/apkbuild
