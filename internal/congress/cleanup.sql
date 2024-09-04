begin transaction;
delete from bill_actions;
delete from vote_records;
delete from votes;
delete from bill_cosponsors;
delete from bills;
commit;