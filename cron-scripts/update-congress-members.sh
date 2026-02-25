#!/usr/bin/env bash

# This script updates the congress members from the unitedstates/congress-legislators GitHub repository.

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

echo "Updating congress members..."

# Define the URL for the JSON data
JSON_URL="https://unitedstates.github.io/congress-legislators/legislators-current.json"

# Create a temporary file to store the downloaded JSON
TEMP_FILE=$(mktemp /tmp/congress_members_XXXXXX.json)

# Download the JSON file using wget
echo "Downloading JSON from $JSON_URL to $TEMP_FILE..."
wget -q -O "$TEMP_FILE" "$JSON_URL"

if [ $? -ne 0 ]; then
    echo "Error: Failed to download JSON file from $JSON_URL"
    rm -f "$TEMP_FILE"
    exit 1
fi

echo "JSON file downloaded successfully."

# Execute the Go program to load congress members from the downloaded file
"$BINARY" -load-congress-members -congress-members-file "$TEMP_FILE"

if [ $? -eq 0 ]; then
    echo "Congress members updated successfully."
else
    echo "Error: Failed to update congress members."
    exit 1
fi

# Clean up the temporary file
rm -f "$TEMP_FILE"
