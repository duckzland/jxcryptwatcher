#!/bin/bash

##
## This is the minimal build script that will generate ios binary at /build folder
##
## flags -ldflags "-w -s" -gcflags="-l" is for creating smallest possible file
## 
## Not sure if this will work properly?


# echo "Generating IOS ARM64 binary"
# GOOS=ios \
# GOARCH=arm64 \
# CGO_ENABLED=1 \
# CGO_CFLAGS="-pthread" \
# CGO_LDFLAGS="-pthread" \
# #CC="${clang} --target=aarch64-linux-android26" \
# go build -tags production -ldflags "-w -s" -gcflags="-l" -o build/jxwatcher-ios-arm64 .

# GOOS=ios \
# GOARCH=arm64 \
# CGO_ENABLED=1 \
# CC="$(xcrun --sdk iphoneos --find clang)" \
# CGO_CFLAGS="-pthread" \
# CGO_LDFLAGS="-pthread" \
# go build -tags production \
#   -ldflags="-w -s" \
#   -gcflags="-l" \
#   -o build/jxwatcher-ios-arm64 .

echo "IOS build is not implemented yet"