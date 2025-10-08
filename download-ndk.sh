#!/bin/bash

## ================================================================
## Android NDK Downloader
## ================================================================
##
## Required dependencies:
##   sudo apt install curl unzip
##
## Usage:
##   ./download_ndk.sh <version> [china]
##   Example: ./download_ndk.sh r29 china
##
## Notes:
## - If 'china' is passed as second argument, Tencent mirror is used.
## - Extracts using unzip quietly.
##
## ================================================================

set -e

echo_error() {
  echo -e "\033[0;31m- $1\033[0m"
}

echo_success() {
  echo -e "\033[0;32m- $1\033[0m"
}

echo_start() {
  echo -e "\033[1m$1\033[0m"
}

NDK_VERSION="${1:-r29}"
MIRROR="${2:-default}"

# Check for curl
if ! command -v curl >/dev/null 2>&1; then
  echo_error "curl is not installed."
  echo "Please install it with: sudo apt update && sudo apt install curl"
  exit 1
fi

# Check for unzip
if ! command -v unzip >/dev/null 2>&1; then
  echo_error "unzip is not installed."
  echo "Please install it with: sudo apt update && sudo apt install unzip"
  exit 1
fi

# Choose base URL
if [ "$MIRROR" == "china" ]; then
  NDK_BASE_URL="https://mirrors.cloud.tencent.com/AndroidSDK"
else
  NDK_BASE_URL="https://dl.google.com/android/repository"
fi

NDK_FILENAME="android-ndk-${NDK_VERSION}-linux.zip"
DOWNLOAD_DIR="build/libs"
NDK_URL="${NDK_BASE_URL}/${NDK_FILENAME}"

mkdir -p "$DOWNLOAD_DIR"

echo_start "Downloading NDK version $NDK_VERSION from $NDK_URL..."
curl -s -C - -L -o "${DOWNLOAD_DIR}/${NDK_FILENAME}" "$NDK_URL"

if [ -f "${DOWNLOAD_DIR}/${NDK_FILENAME}" ]; then
  echo_success "Download complete: ${DOWNLOAD_DIR}/${NDK_FILENAME}"
else
  echo_error "Download failed."
  exit 1
fi

echo_start "Extracting NDK..."
unzip -qq -o "${DOWNLOAD_DIR}/${NDK_FILENAME}" -d "${DOWNLOAD_DIR}"
echo_success "NDK extracted to ${DOWNLOAD_DIR}/android-ndk-${NDK_VERSION}"