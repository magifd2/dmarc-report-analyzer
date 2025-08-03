#!/bin/bash

# Ensure we are in the frontend directory
cd "$(dirname "$0")" || exit

PID_FILE="server.pid"
LOG_FILE="server.log"

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null; then
        echo "Frontend server is already running with PID $PID."
        exit 0
    else
        echo "Stale PID file found. Removing $PID_FILE."
        rm "$PID_FILE"
    fi
fi

echo "Starting frontend server..."

# Run npm dev in background and capture its output to a temporary log file first
npm run dev > "$LOG_FILE.tmp" 2>&1 &
NPM_DEV_PID=$!

# Give the server a moment to start and bind to the port
sleep 5

# Read the temporary log file to find the actual port Vite is using
# Using standard grep and sed for broader compatibility
ACTUAL_PORT=$(grep -o 'http://localhost:[0-9]\+' "$LOG_FILE.tmp" | head -1 | sed 's/http:\/\/localhost://')

if [ -z "$ACTUAL_PORT" ]; then
    echo "Error: Could not determine actual port used by Vite. Check $LOG_FILE.tmp"
    cat "$LOG_FILE.tmp" >> "$LOG_FILE"
    rm "$LOG_FILE.tmp"
    exit 1
fi

# Find the PID of the process listening on the actual port
SERVER_PID=$(lsof -t -i :"$ACTUAL_PORT" | head -1)

if [ -z "$SERVER_PID" ]; then
    echo "Error: Could not find frontend server PID listening on port $ACTUAL_PORT."
    echo "Please check $LOG_FILE.tmp for details."
    cat "$LOG_FILE.tmp" >> "$LOG_FILE"
    rm "$LOG_FILE.tmp"
    exit 1
fi

echo "$SERVER_PID" > "$PID_FILE"

# Append the temporary log to the main log file
cat "$LOG_FILE.tmp" >> "$LOG_FILE"
rm "$LOG_FILE.tmp"

echo "Frontend server started with PID $SERVER_PID on port $ACTUAL_PORT. Log file: $LOG_FILE"
echo "To stop the server, run: ./stop.sh"