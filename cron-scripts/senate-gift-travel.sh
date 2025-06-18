#!/usr/bin/env bash
# Downloads the travel file from the US Senate Website

# https://www.senate.gov/pagelayout/legislative/g_three_sections_with_teasers/lobbyingdisc.htm



pwd
filePath=$(mktemp)
BINARY=~final_stockbot/the_final_stockbot/the_final_stockbot

wget -O "$filePath" https://giftrule-disclosure.senate.gov/media/giftruledownloads/giftruledata.zip
if [[ ! -x "$BINARY" ]]; then

    if [[ ! -f "$BINARY" ]]; then
        echo "Error: the_final_stockbot binary not found."
    else
        "$BINARY" -script senate-travel -file $filePath
    fi

else
    $BINARY -script senate-travel -file $filePath
fi
rm $filePath