#!/bin/bash

# Ensure we are in the backend directory
cd "$(dirname "$0")" || exit 1

# Define paths
APP_NAME="dmarc-report-analyzer-backend" # Name of the compiled executable
PID_FILE="server.pid"
LOG_FILE="server.log"
BUILD_DIR="bin" # Directory for compiled executable

# Check if the server is already running
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null; then
        echo "Server is already running with PID $PID."
        exit 0
    else
        echo "Stale PID file found. Removing $PID_FILE."
        rm "$PID_FILE"
    fi
fi

echo "Building Go backend..."
# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"
# Build the Go application
go build -o "$BUILD_DIR/$APP_NAME" src/main.go
if [ $? -ne 0 ]; then
    echo "Go build failed."
    exit 1
fi # Added missing fi
echo "Build successful: $BUILD_DIR/$APP_NAME"

echo "Starting server..."
# Set JWT_SECRET for the nohup process. REPLACE THIS WITH YOUR ACTUAL SECRET.
export JWT_SECRET="your_super_secret_jwt_key_at_least_32_chars_long"

nohup "$BUILD_DIR/$APP_NAME" > "$LOG_FILE" 2>&1 &
PID=$!
echo "$PID" > "$PID_FILE"

echo "Server started with PID $PID. Log file: $LOG_FILE"
echo "To stop the server, run: ./stop.sh"