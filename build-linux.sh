#!/bin/sh

##
## This is the minimal build script that will generate linux binary at /build folder
##
## flags -ldflags "-w -s" -gcflags="-l" is for creating smallest possible file
## 


echo "Generating Linux binary"
go build -ldflags "-w -s" -gcflags="-l" -o build/jxwatcher src/*