SELECT 
    registrant_name, payee_name, honoree_name,
     SUM(amount) as Amount, 
     Count(*) as Count 
FROM contributions_etl 
WHERE 
    filing_year = ? AND contribution_type = ? 
GROUP BY registrant_name, payee_name, honoree_name
 ORDER BY Amount DESC, registrant_name DESC LIMIT 50