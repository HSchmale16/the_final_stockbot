#!/usr/bin/env bash

# This script updates the congressional hearings by fetching the latest RSS feed from GovInfo.

# Navigate to the application directory
cd "$(dirname "$0")/../"

# Define the path to the binary
BINARY="./the_final_stockbot"

# Check if the binary exists and is executable
if [[ ! -x "$BINARY" ]]; then
    echo "Error: the_final_stockbot binary not found or not executable at $BINARY"
    # Attempt to build the binary if it's not found.
    echo "Attempting to build the binary..."
    go build -o the_final_stockbot .
    if [[ ! -x "$BINARY" ]]; then
        echo "Error: Failed to build the_final_stockbot binary."
        exit 1
    fi
fi

echo "Loading congressional hearings..."

# Execute the Go program to run the hearing fetcher service
"$BINARY" --run-hearing-fetcher

if [ $? -eq 0 ]; then
    echo "Congressional hearings updated successfully."
else
    echo "Error: Failed to update congressional hearings."
    exit 1
fi
