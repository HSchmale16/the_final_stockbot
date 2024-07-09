WITH F AS (
SELECT
    filing_year
    , json_item->>'filing_type'
    , json_item->>'income'
    , json_item->>'expenses'
    , json_item->'registrant'->>'name'
    , json_item->'client'->>'name' client
    , value->>'general_issue_code_display'
    , value->>'description'
    , value->>'foreign_entity_issues' AS foreign_entity_issues
FROM filings, json_each(filings.json_item, '$.lobbying_activities')
)
SELECT client, foreign_entity_issues FROM F
WHERE foreign_entity_issues != ''