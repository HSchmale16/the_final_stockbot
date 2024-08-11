#!/usr/bin/env bash
# Downloads the travel file from the US House Website

pwd
wget https://disclosures-clerk.house.gov/public_disc/gift-pdfs/2024Travel.zip
./the_final_stockbot/the_final_stockbot -script house-travel -file 2024Travel.zip
rm 2024Travel.zip