WITH unfolded AS (
select
    json_extract(json_item, '$.registrant.name') as registrant_name
    , json_item->>'filing_year' as filing_year
    , value->>'amount' as amount
    , value->>'honoree_name' as honoree_name
    , value->>'contribution_type' as contribution_type
    , value->>'payee_name' as payee_name
    , value->>'date' as contribution_date
from contributions, json_each(contributions.json_item, '$.contribution_items') 
where 
    json_item->>'no_contributions' = false
)
select 
    registrant_name
    , printf('%,.2f', sum(cast(amount as float))) as total_amount
    , count(*)
from unfolded
group by registrant_name