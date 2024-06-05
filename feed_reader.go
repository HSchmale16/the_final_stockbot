package main

import (
	"log"
	"time"
	_ "fmt"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
)



func findUnfetchedFeeds(db *gorm.DB) ([]RSSFeed, error) {
	var feeds []RSSFeed
	err := db.Where("last_fetched < ?", time.Now().Add(-15*time.Minute)).Find(&feeds).Error
	if err != nil {
		log.Println("Failed to find unfetched feeds:", err)
		return nil, err
	}
	return feeds, nil
}

func loadFeed(db *gorm.DB, feed *RSSFeed) {
	// Use gofeed to fetch the url
	fp := gofeed.NewParser()
	parsedFeed, _ := fp.ParseURL(feed.Link) // Use the URL from the feed parameter
	log.Println(parsedFeed.Title)

	for _, item := range parsedFeed.Items {
		// Process each feed item
		// Example: log.Println(item.Title)
		// Insert the feed item into the database
		pubDate, err := time.Parse(time.RFC1123Z, item.Published)
		if err != nil {
			log.Println("Failed to parse pubDate:", err)
			continue
		}

		rssItem := &RSSItem{
			GUID:        item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			PubDate:     pubDate,
			FeedID:      feed.ID,
		}

		
		rssItem.PubDate = pubDate
		// Check if the RSS item already exists in the database
		var existingItem RSSItem
		err = db.Where("guid = ?", rssItem.GUID).First(&existingItem).Error
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			log.Println("Failed to check existing item:", err)
			continue
		}
		if existingItem.ID != 0 {
			// RSS item already exists, skip inserting
			continue
		}

		// Insert the feed item into the database
		err = db.Create(rssItem).Error
		if err != nil {
			log.Println("Failed to insert feed item:", err)
			continue
		}
	}

	// Update the feed with the current time
	feed.LastFetched = time.Now()
	db.Save(feed)
}



func fetchFeeds(db *gorm.DB) {
	feeds, err := findUnfetchedFeeds(db)
	if err != nil {
		return
	}
	for _, feed := range feeds {
		loadFeed(db, &feed)
	}
}