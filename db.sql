CREATE TABLE contributions (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    filing_year INT GENERATED ALWAYS AS (json_item->>'filing_year') STORED
);

CREATE TABLE filings (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL
);