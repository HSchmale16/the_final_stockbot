SELECT tag_id
	, (SELECT name FROM tag where id = tag_id)
	, COUNT(*) 
from govt_rss_item_tag
group by tag_id
order by COUNT(*) DESC

SELECT title, pub_date
	, (SELECT COUNT(*) FROM govt_rss_item_tag WHERE govt_rss_item_id = govt_rss_item.id) as tag_count
FROM govt_rss_item 
order by tag_count DESC

