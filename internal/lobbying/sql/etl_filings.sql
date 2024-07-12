CREATE TABLE filings_etl AS
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
    LEFT JOIN json_each(filings.json_item, '$.foreign_entities') as foreign_entities
