#!/usr/bin/env bash
# Downloads the travel file from the US Senate Website

# https://www.senate.gov/pagelayout/legislative/g_three_sections_with_teasers/lobbyingdisc.htm



pwd
wget -O giftruledata.zip https://giftrule-disclosure.senate.gov/media/giftruledownloads/giftruledata.zip
if [[ ! -x ./the_final_stockbot/the_final_stockbot ]]; then

    if [[ ! -f ./the_final_stockbot ]]; then
        echo "Error: the_final_stockbot binary not found."
    else
        ./the_final_stockbot -script senate-travel -file giftruledata.zip
    fi

else
    ./the_final_stockbot/the_final_stockbot -script senate-travel -file giftruledata.zip
fi
rm giftruledata.zip