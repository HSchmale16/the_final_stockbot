#!/usr/bin/env bash

# This script updates the congress members from the unitedstates/congress-legislators GitHub repository.
# Both the current and historical legislator feeds are fetched automatically by the binary.

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

echo "Updating congress members (current + historical)..."

# The binary fetches both legislators-current.json and legislators-historical.json internally.
"$BINARY" -load-congress-members

if [ $? -eq 0 ]; then
    echo "Congress members updated successfully."
else
    echo "Error: Failed to update congress members."
    exit 1
fi
