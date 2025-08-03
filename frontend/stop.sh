#!/bin/bash

# Ensure we are in the frontend directory
cd "$(dirname "$0")" || exit

PID_FILE="server.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "PID file ($PID_FILE) not found. Server may not be running."
    exit 0
fi

PID=$(cat "$PID_FILE")

if ps -p "$PID" > /dev/null; then
    echo "Stopping frontend server with PID $PID..."
    kill "$PID"
    # Wait for the process to terminate
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null; then
            echo "Server stopped."
            rm "$PID_FILE"
            exit 0
        fi
        sleep 1
    done
    echo "Failed to stop server with PID $PID. Manual intervention may be required."
    exit 1
else
    echo "Server with PID $PID> is not running. Removing stale PID file."
    rm "$PID_FILE"
    exit 0
fi
