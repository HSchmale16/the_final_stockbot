-- Find the sponsor of each item
-- Then create a weighting between them
WITH sponsors AS
(
    SELECT db_congress_member_bio_guide_id, govt_rss_item_id FROM congress_member_sponsored WHERE role = 'SPONSOR' AND chamber = ?1
),
cosponsors AS (
    SELECT db_congress_member_bio_guide_id, govt_rss_item_id FROM congress_member_sponsored WHERE role = 'COSPONSOR' AND chamber = ?1
),
sponsor_weights AS (
    SELECT 
        sponsors.db_congress_member_bio_guide_id AS sponsor_id,
        cosponsors.db_congress_member_bio_guide_id AS cosponsor_id,
        2 as weight
    FROM sponsors
    INNER JOIN cosponsors ON sponsors.govt_rss_item_id = cosponsors.govt_rss_item_id
    UNION ALL
    SELECT 
        cosponsors.db_congress_member_bio_guide_id AS cosponsor_id,
        sponsors.db_congress_member_bio_guide_id AS sponsor_id,
        1 as weight
    FROM sponsors
    INNER JOIN cosponsors ON sponsors.govt_rss_item_id = cosponsors.govt_rss_item_id
)
SELECT sponsor_id as source, cosponsor_id as target, SUM(weight) as value
FROM sponsor_weights
GROUP BY sponsor_id, cosponsor_id
ORDER BY weight DESC