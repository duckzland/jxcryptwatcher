#!/bin/sh


## 
## You must install mingw for compiling go to windows binary
## sudo apt install gcc-mingw-w64
##

echo "Generating Windows binary"

## Setup the flags
## -pthread will speed up the compile time
## -w -s will create smallest file possible
## -H=windowsgui is needed to fix windows showing terminal eventhough this is a GUI program
GOOS=windows \
GOARCH=amd64  \
CGO_ENABLED=1 \
CGO_CFLAGS="-pthread" \
CGO_LDFLAGS="-pthread" \
CC=/usr/bin/x86_64-w64-mingw32-gcc \
go build -ldflags "-w -s -H=windowsgui" -o build/jxwatcher.exe src/*
