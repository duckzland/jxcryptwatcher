#!/bin/bash


## 
## You must install android ndk for compiling go to android arm64 binary
## sudo apt install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu google-android-ndk-r26c-installer
##
## Make sure the following paths and versions are correct:
## - Sysroot exists:
##   /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot
##
## - Target lib directory for API level 26 exists:
##   /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android/26
##
## - Clang binary is available:
##   /usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang
##
## - The API level used in '--target=aarch64-linux-android26' matches the lib directory version:
##   '26' in both the --target flag and the sysroot path

echo "Generating Android arm64 binary"

## Setup the flags
## -pthread will speed up the compile time
## -w -s will create smallest file possible

## This is needed to overcome the doublequote problem
extldflags='-extldflags="-B/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android/26 -fuse-ld=lld -Wl,--gc-sections"'

GOOS=android \
GOARCH=arm64 \
CGO_ENABLED=1 \
CGO_CFLAGS="--sysroot=/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot -pthread" \
CGO_LDFLAGS="-L/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/lib/aarch64-linux-android/26 -pthread" \
CC="/usr/lib/android-ndk/toolchains/llvm/prebuilt/linux-x86_64/bin/clang --target=aarch64-linux-android26" \
go build -tags production \
-ldflags="-w -s -linkmode=external $extdflags" \
-o build/jxwatcher-android-arm64 .