#!/bin/bash
# Usage: ./debug-pprof_live.sh heap
# Supported profiles: heap, allocs, goroutine, block, etc.
# Please enable PProfDebug in build_debug.go before using.
# Install pprof if not already installed: go install github.com/google/pprof@latest

PROFILE=${1:-heap}
URL="http://localhost:6060/debug/pprof/$PROFILE"

echo "Fetching profile from $URL..."
TMPFILE=$(mktemp /tmp/pprof-${PROFILE}-XXXXXX.pb.gz)

# Fetch the profile data
curl -s "$URL" -o "$TMPFILE"

echo "Launching pprof web viewer on http://localhost:8080..."
go tool pprof -http=:8080 "$TMPFILE"

echo "Cleaning up temporary profile: $TMPFILE"
rm -f "$TMPFILE"