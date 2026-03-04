#!/usr/bin/env bash
set -e

DUMP=$(find /remote_data_mount -name "congress*.sql" | sort | tail -n1)
if [ -z "$DUMP" ]; then
  echo "No dump file found in /remote_data_mount — skipping restore."
  exit 0
fi

echo "Loading dump: $DUMP"
psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$DUMP"
echo "Dump loaded successfully."
