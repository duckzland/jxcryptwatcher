#!/bin/sh


## This is the minimal build script that will generate linux binary at /build folder


cd ./build

echo "Generating Standard binary"
go build -o jxwatcher ../src/*


cd ../