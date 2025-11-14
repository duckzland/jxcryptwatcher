#!/bin/bash

## ================================================================
## JXWatcher Pre-Build Script
## ================================================================
## This script performs pre-build tasks:
##   1. Verifies Git is installed and creates a temporary branch.
##   2. Strips all JC.{function} calls from Go source files.
##   3. Search for notification string and create const for replacement
##
## Requirement:
## 1. Git - apt install git
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

mode="$1"
live_mode=false
if [ "$mode" == "live" ]; then
  live_mode=true
  echo_success "Running in LIVE mode — skipping Git operations"
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

if [ "$live_mode" = false ]; then

  if ! command -v git &> /dev/null; then
    echo_error "Git is not installed. Please install Git before proceeding."
    exit 1
  fi

  if ! git diff --quiet || ! git diff --cached --quiet; then
    echo_error "Uncommitted changes detected. Please commit or stash them before proceeding."
    exit 1
  fi

  if git show-ref --verify --quiet refs/heads/"$temp_branch"; then
    echo_success "Switching to existing temporary branch: $temp_branch"
    git checkout "$temp_branch"
  else
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    echo_success "Creating temporary branch: $temp_branch from $current_branch"
    git checkout -b "$temp_branch"
  fi
fi

for func in "${functions[@]}"; do
  find . -type f -name "*.go" | while read -r file; do
    awk -v f="$func" '
      $0 ~ "^[[:space:]]*func[[:space:]]+" f "\\s*\\(" { print; next }
      $0 ~ "JC\\." f "\\s*\\(" {
        print "if false {"
        print $0
        print "}"
        next
      }
      $0 ~ "(^|[^a-zA-Z0-9_])" f "\\s*\\(" {
        print "if false {"
        print $0
        print "}"
        next
      }
      { print }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
  done

  echo_success "Debug function: ${func} calls neutralized"
done

if [ "$live_mode" = false ]; then
  git add .
  git commit -m "$commit_msg"
  echo_success "Changes committed to $temp_branch"
fi

echo_success "Pre-build steps completed successfully."