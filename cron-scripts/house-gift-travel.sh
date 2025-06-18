#!/usr/bin/env bash
# Downloads the travel file from the US House Website

pwd
YEAR=$(date +"%Y")
if [ -n "$1" ]; then
    YEAR=$1
fi

filePath=$(mktemp)

wget -O $filePath https://disclosures-clerk.house.gov/public_disc/gift-pdfs/${YEAR}Travel.zip
BINARY=~final_stockbot/the_final_stockbot/the_final_stockbot

if [[ ! -x "$BINARY" ]]; then

    if [[ ! -f ./the_final_stockbot ]]; then
        echo "Error: the_final_stockbot binary not found."
    else
        "$BINARY" -script house-travel -file $filePath
    fi

else
    "$BINARY" -script house-travel -file $filePath
fi
rm $filePath