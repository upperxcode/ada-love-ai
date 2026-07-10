#!/bin/bash
# Script to analyze context logs from Ada Love AI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default log path
DEFAULT_LOG_PATH="$HOME/.config/ada-love/context_logs.jsonl"

# Check if log file exists
if [ -f "$DEFAULT_LOG_PATH" ]; then
    LOG_PATH="$DEFAULT_LOG_PATH"
else
    echo "Usage: $0 [log-file-path]"
    echo ""
    echo "Default log path: $DEFAULT_LOG_PATH"
    echo ""
    echo "If no log file exists, run Ada Love AI and send some messages first."
    exit 1
fi

echo "Analyzing context logs from: $LOG_PATH"
echo ""

cd "$PROJECT_ROOT"
go run ./scripts/analyze-contexts.go "$LOG_PATH"