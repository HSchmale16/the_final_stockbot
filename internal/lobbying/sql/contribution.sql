WITH unfolded AS (
select
    CAST(json_extract(json_item, '$.registrant.name') AS VARCHAR) as registrant_name
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
group by filing_year, registrant_name
order by sum(cast(amount as float)) DESC LIMIT 10;
