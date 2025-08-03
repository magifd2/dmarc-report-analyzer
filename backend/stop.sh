#!/bin/bash

# Ensure we are in the backend directory
cd "$(dirname "$0")" || exit 1

PID_FILE="server.pid"

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null; then
        echo "Stopping server with PID $PID..."
        kill "$PID"
        # Wait a bit for the process to terminate
        sleep 2
        if ps -p "$PID" > /dev/null; then
            echo "Server did not stop gracefully. Force killing PID $PID..."
            kill -9 "$PID"
        fi
    else
        echo "Server process with PID $PID not found."
    fi
    rm "$PID_FILE"
    echo "Server stopped and PID file removed."
else
    echo "PID file ($PID_FILE) not found. Server may not be running."
fi
