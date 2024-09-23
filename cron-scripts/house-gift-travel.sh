#!/usr/bin/env bash
# Downloads the travel file from the US House Website

pwd
YEAR=$(date +"%Y")
if [ -n "$1" ]; then
    YEAR=$1
fi
wget https://disclosures-clerk.house.gov/public_disc/gift-pdfs/${YEAR}Travel.zip
./the_final_stockbot/the_final_stockbot -script house-travel -file ${YEAR}Travel.zip
rm ${YEAR}Travel.zip