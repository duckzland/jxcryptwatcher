#!/bin/bash
# Usage: ./debug-gowatcher.sh [gops_command] [process_keyword]
# Default command: memstats
# Default keyword: jxwatcher
# Please enable GopsDebug in build_debug.go before using.
# Install gops if not already installed: go install github.com/google/gops@latest

CMD=${1:-memstats}
KEYWORD=${2:-jxwatcher}

# Find the PID of the process matching the keyword
PID=$(pgrep -f "$KEYWORD")

if [ -z "$PID" ]; then
  echo "Error: No process found matching '$KEYWORD'"
  exit 1
fi

echo "Found process '$KEYWORD' with PID: $PID"
echo "Running: gops $CMD $PID (refresh every 1s)"
echo "Press Ctrl+C to stop."

# Run gops command every second
watch -n 1 "gops $CMD $PID"