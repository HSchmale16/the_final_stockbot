-- Just a file to help copilot with column names and tables since it's not able to read the schema from the database
-- and the table names got named stupid while I was hurrying.
CREATE TABLE `db_committee_memberships` (
    `created_at` datetime,
    `updated_at` datetime,
    `db_congress_member_bio_guide_id` text,
    `db_congress_committee_thomas_id` text,
    `rank` integer,
    `title` text,
PRIMARY KEY (`db_congress_member_bio_guide_id`,`db_congress_committee_thomas_id`),CONSTRAINT `fk_congress_member_committees` FOREIGN KEY (`db_congress_member_bio_guide_id`) REFERENCES `congress_member`(`bio_guide_id`),CONSTRAINT `fk_congress_committee_memberships` FOREIGN KEY (`db_congress_committee_thomas_id`) REFERENCES `congress_committee`(`thomas_id`))