WITH sponsors AS
(
    SELECT db_congress_member_bio_guide_id, govt_rss_item_id FROM congress_member_sponsored WHERE role = 'SPONSOR' AND chamber = 'H'
),
cosponsors AS (
    SELECT db_congress_member_bio_guide_id, govt_rss_item_id FROM congress_member_sponsored WHERE role = 'COSPONSOR' AND chamber = 'H'
),
sponsor_cosponsor_connections AS (
    SELECT 
        s.db_congress_member_bio_guide_id AS sponsor_id,
        c.db_congress_member_bio_guide_id AS cosponsor_id,
        s.govt_rss_item_id AS bill_id
    FROM sponsors s
    JOIN cosponsors c ON s.govt_rss_item_id = c.govt_rss_item_id
    WHERE s.db_congress_member_bio_guide_id != c.db_congress_member_bio_guide_id AND
        -- identify those cosponsors who don't serve on any of the same committees
        NOT EXISTS (
            SELECT 1
            FROM db_committee_memberships s_committees
            JOIN db_committee_memberships c_committees
            ON s_committees.db_congress_member_bio_guide_id = s.db_congress_member_bio_guide_id
            WHERE
                s_committees.db_congress_committee_thomas_id = c_committees.db_congress_committee_thomas_id
        )
)
SELECT * FROM sponsor_cosponsor_connections