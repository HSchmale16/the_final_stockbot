package main

import (
	"errors"
	_ "fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

type DocumentGrabber func(*goquery.Document) string

func getGrabMethod(link string) DocumentGrabber {
	contentGrabMethods := map[string]DocumentGrabber{
		"pr.com": processPR_com_HTML,
	}

	for domain, method := range contentGrabMethods {
		if strings.Contains(link, domain) {
			return method
		}
	}
	return nil
}

func findUnfetchedFeeds(db *gorm.DB) ([]RSSFeed, error) {
	var feeds []RSSFeed
	err := db.Where("last_fetched < ?", time.Now().Add(-15*time.Minute)).Find(&feeds).Error
	if err != nil {
		log.Println("Failed to find unfetched feeds:", err)
		return nil, err
	}
	return feeds, nil
}

func processPR_com_HTML(doc *goquery.Document) string {
	// Extract the desired selector
	selector := "body > main > div > article > article.press-release__content > section.press-release__body"
	selectedContent := doc.Find(selector)

	// Remove style nodes from the selected content
	selectedContent.Find("style").Remove()

	return selectedContent.Text()
}

func loadFeed(db *gorm.DB, feed *RSSFeed) []RSSItem {
	// Use gofeed to fetch the url
	fp := gofeed.NewParser()
	parsedFeed, _ := fp.ParseURL(feed.Link) // Use the URL from the feed parameter
	log.Println("Reading feed", feed.Title)

	newItems := []RSSItem{}

	// Function pointer that maps a string to a specific function
	grabMethod := getGrabMethod(feed.Link)

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
			Guid:        item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			PubDate:     pubDate,
			FeedID:      feed.ID,
		}

		// Check if the RSS item already exists in the database
		var existingItem RSSItem
		err = db.Where("guid = ?", rssItem.Guid).First(&existingItem).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Failed to check existing item:", err)
			continue
		}
		if existingItem.ID != 0 {
			// RSS item already exists, skip inserting
			continue
		}

		if grabMethod != nil {
			// Download the link as HTML
			log.Println("Getting HTML for", rssItem.Link)
			resp, err := http.Get(rssItem.Link)
			if err != nil {
				log.Println("Failed to download HTML:", err)
				continue
			}
			defer resp.Body.Close()

			log.Println("Parsing HTML for", rssItem.Link)
			// Parse the HTML document
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Println("Failed to parse HTML:", err)
				continue
			}

			content := grabMethod(doc)
			rssItem.ArticleBody = &content
		}

		// Insert the feed item into the database
		err = db.Create(rssItem).Error
		if err != nil {
			log.Println("Failed to insert feed item:", err)
			continue
		}
		newItems = append(newItems, *rssItem)
	}

	// Update the feed with the current time
	feed.LastFetched = time.Now()
	db.Save(feed)

	return newItems
}

func fetchFeeds(db *gorm.DB) {
	feeds, err := findUnfetchedFeeds(db)
	if err != nil {
		return
	}
	for _, feed := range feeds {
		newItems := loadFeed(db, &feed)
		log.Print("Fetched ", len(newItems), " new items from ", feed.Link)
	}
}
