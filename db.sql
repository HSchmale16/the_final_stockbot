CREATE TABLE contributions (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL
);

CREATE TABLE filings (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL
);