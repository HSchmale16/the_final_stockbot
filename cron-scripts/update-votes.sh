#!/usr/bin/env bash

# Navigate to the application directory
cd "$(dirname "$0")/../"

# Define the path to the binary
BINARY="./the_final_stockbot"

# Check if the binary exists and is executable
if [[ ! -x "$BINARY" ]]; then
    echo "Error: the_final_stockbot binary not found or not executable at $BINARY"
    echo "Attempting to build the binary..."
    go build -o the_final_stockbot .
    if [[ ! -x "$BINARY" ]]; then
        echo "Error: Failed to build the_final_stockbot binary."
        exit 1
    fi
fi

echo "Loading votes from UCLA Voteview..."
"$BINARY" -script load-voteview 119

echo "Backfilling last vote dates..."
"$BINARY" -script backfill-last-vote-date

echo "Votes update completed successfully."
