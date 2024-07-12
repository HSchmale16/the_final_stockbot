ATTACH 'contribution_list.db' (TYPE SQLITE);

DROP TABLE IF EXISTS contributions_etl;
CREATE TABLE contributions_etl AS
SELECT 
    "uuid"
    , cast(registrant_name as text) as registrant_name
    , cast(filing_year as int) as filing_year
    , cast(amount as float) as amount
    , cast(honoree_name as text) as honoree_name
    , cast(contribution_type as text) as contribution_type
    , cast(payee_name as text) as payee_name
    , cast(contribution_date as text) as contribution_date
FROM contribution_list.lobbyist_contributions;

DROP TABLE IF EXISTS filings_etl;
CREATE TABLE filings_etl AS
SELECT * FROM contribution_list.filings_etl;