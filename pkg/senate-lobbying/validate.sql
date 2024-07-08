SELECT 
FROM filings, json_each(filings.json_item, '$.lobbying_activities') 
WHERE
    json_item->>