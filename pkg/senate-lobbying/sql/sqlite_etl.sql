.timer on
.eqp on
.shell echo DROPING AND CREATING filings_etl
DROP TABLE IF EXISTS filings_etl;
CREATE TABLE filings_etl (
    uuid,
    filing_year INT,
    filing_type TEXT,
    income TEXT,
    expenses TEXT,
    reg_name TEXT,
    client TEXT,
    issue_code_display TEXT,
    issue_description TEXT,
    foreign_entity_issues TEXT,
    foreign_entity_name TEXT,
    foreign_entity_country TEXT,
    foreign_entity_contribution TEXT,
    foreign_entity_ownership_percentage TEXT
);

INSERT INTO filings_etl
SELECT
    CAST(uuid AS TEXT) as uuid
    , filing_year
    , json_item->>'filing_type' as filing_type
    , json_item->>'income' as income 
    , json_item->>'expenses' as expenses
    , json_item->'registrant'->>'name' as reg_name
    , json_item->'client'->>'name' client
    , lobbying.value->>'general_issue_code_display' as issue_code_display
    , lobbying.value->>'description' as issue_description
    , lobbying.value->>'foreign_entity_issues' AS foreign_entity_issues
    , foreign_entities.value->>'name' AS foreign_entity_name
    , foreign_entities.value->>'country' AS foreign_entity_country
    , foreign_entities.value->>'contribution' as foreign_entity_contribution
    , foreign_entities.value->>'ownership_percentage' as foreign_entity_ownership_percentage
FROM 
    filings, 
    json_each(filings.json_item, '$.lobbying_activities') as lobbying
    LEFT JOIN json_each(filings.json_item, '$.foreign_entities') as foreign_entities;

.shell echo DROPING AND CREATING lobbyist_contributions

DROP TABLE IF EXISTS lobbyist_contributions;
CREATE TABLE lobbyist_contributions (
    uuid TEXT,
    registrant_name TEXT,
    filing_year INT,
    amount float,
    contribution_type TEXT,
    honoree_name TEXT,
    payee_name TEXT,
    contribution_date TEXT
);


INSERT INTO lobbyist_contributions
select
    cast(uuid as text) as uuid
    , json_extract(json_item, '$.registrant.name') as registrant_name
    , json_item->>'filing_year' as filing_year
    , value->>'amount' as amount
    , value->>'contribution_type' as contribution_type
    , value->>'honoree_name' as honoree_name
    , value->>'payee_name' as payee_name
    , value->>'date' as contribution_date
from contributions, json_each(contributions.json_item, '$.contribution_items') 
where 
    json_item->>'no_contributions' = false;
