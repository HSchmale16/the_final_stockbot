CREATE TABLE contributions (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    filing_year INT GENERATED ALWAYS AS (json_item->>'filing_year') STORED
    update_count INT DEFAULT 0
);

CREATE TRIGGER trg_contributions AFTER UPDATE ON contributions
    FOR EACH ROW WHEN new.json_item <> old.json_item
    BEGIN 
         UPDATE contributions SET update_count = update_count + 1 WHERE uuid = new.uuid;
    END;

CREATE TABLE filings (
    uuid STRING PRIMARY KEY,
    json_item JSONB NOT NULL
);