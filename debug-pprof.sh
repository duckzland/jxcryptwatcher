#!/bin/bash
# Usage: ./debug-pprof_live.sh [profile]
# Supported profiles: heap, allocs, goroutine, block, etc.
# Please enable PProfDebug in build_debug.go before using.
# Install pprof if not already installed: go install github.com/google/pprof@latest

PROFILE=${1:-heap}
URL="http://localhost:6060/debug/pprof/$PROFILE"

# Check required tools
for tool in go curl; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "Error: '$tool' is not installed. Please install it and try again."
    exit 1
  fi
done

# Check if pprof is installed via go
if ! go tool pprof -h >/dev/null 2>&1; then
  echo "Error: 'pprof' tool not available via 'go'. Please run:"
  echo "  go install github.com/google/pprof@latest"
  exit 1
fi

echo "Fetching profile from $URL..."
TMPFILE=$(mktemp /tmp/pprof-${PROFILE}-XXXXXX.pb.gz)

# Fetch the profile data
curl -s "$URL" -o "$TMPFILE"

echo "Launching pprof web viewer on http://localhost:8080..."
go tool pprof -http=:8080 "$TMPFILE"

echo "Cleaning up temporary profile: $TMPFILE"
rm -f "$TMPFILE"