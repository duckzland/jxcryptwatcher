#!/bin/sh


## 
## You must install windows additional build packet for ubuntu
## apt install gcc-mingw-w64-x86-64-win32
## 
## or for windows 386
## apt install gcc-mingw-w64-i686-win32
## 
## TODO: Figure out why windows build always firing console?
## TODO: why -ldflags="-H=windowsgui" not honored?

cd ./build

echo "Generating Windows binary"
CGO_ENABLED=1 CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_LDFLAGS="-static-libgcc -static -H windowsgui" CGO_CFLAGS="-pthread" CGO_LDFLAGS="-pthread" go build -ldflags "-H windowsgui" -o jxwatcher.exe ../src/*

cd ../