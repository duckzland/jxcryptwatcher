#!/bin/bash

## ================================================================
## JXWatcher Pre-Build Script
## ================================================================
## This script performs pre-build tasks:
##   1. Verifies Git is installed and creates a temporary branch.
##   2. Strips all JC.{function} calls from Go source files.
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

echo_start "Starting pre-build cleanup..."

if ! command -v git &> /dev/null; then
  echo_error "Git is not installed. Please install Git before proceeding."
  exit 1
fi

if ! command -v sed &> /dev/null; then
  echo_error "sed is not installed. Please install sed before proceeding."
  exit 1
fi

if ! command -v find &> /dev/null; then
  echo_error "find is not installed. Please install find before proceeding."
  exit 1
fi

if [ ! -f version.txt ]; then
  echo_error "version.txt not found. Please create one with format 'version=1.0.0'"
  exit 1
fi

temp_branch="temp-build"
functions=("InitLogger" "Logln" "Logf" "PrintMemUsage" "PrintExecTime" "PrintPerfStats")
version=$(grep '^version=' version.txt | cut -d'=' -f2 | tr -d '[:space:]')
timestamp=$(date +"%Y%m%d-%H%M%S")
commit_msg="prebuild cleanup v${version}-${timestamp}"

if git show-ref --verify --quiet refs/heads/"$temp_branch"; then
  echo_success "Switching to existing temporary branch: $temp_branch"
  git checkout "$temp_branch"
else
  current_branch=$(git rev-parse --abbrev-ref HEAD)
  echo_success "Creating temporary branch: $temp_branch from $current_branch"
  git checkout -b "$temp_branch"
fi

# for func in "${functions[@]}"; do
#   find . -type f -name "*.go" -exec sed -i "/JC\.${func}.*/d" {} +
#   echo_success "JC.${func} calls removed."
# done

for func in "${functions[@]}"; do
  find . -type f -name "*.go" -exec sed -i "/JC\.${func}/s/^/if false {\n/; /JC\.${func}/s/$/\n}/" {} +
  echo_success "JC.${func} calls removed."
done

git add .
git commit -m "$commit_msg"

echo_success "Changes committed to $temp_branch"

echo_success "Pre-build steps completed successfully."