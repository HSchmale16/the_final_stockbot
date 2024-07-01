-- We are performing breadth first search to find out which congress people work together most often

WITH edges AS
(
    SELECT 
        cms1.db_congress_member_bio_guide_id first_congress
        , cms2.db_congress_member_bio_guide_id second_congress
        , count(*) cnt
    FROM congress_member_sponsored cms1
    JOIN congress_member_sponsored cms2 ON cms1.govt_rss_item_id = cms2.govt_rss_item_id 
    GROUP BY first_congress, second_congress
),
