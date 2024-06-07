package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type RSSFeed struct {
	gorm.Model
	Title       string `gorm:"unique_index"`
	Description string
	Link        string
	LastFetched time.Time
}

func (RSSFeed) TableName() string {
	return "rss_feeds"
}

type RSSItem struct {
	gorm.Model
	Guid        string  `gorm:"type:text"`
	Title       string  `gorm:"type:text"`
	Description string  `gorm:"type:text"`
	Link        string  `gorm:"type:text"`
	ArticleBody *string `gorm:"type:text"` // this is the article content
	PubDate     time.Time
	FeedID      uint
	Feed        RSSFeed `gorm:"foreignkey:FeedID"`
}

func (RSSItem) TableName() string {
	return "rss_items"
}

type MarketSecurity struct {
	gorm.Model
	Symbol   string `gorm:"unique_index"`
	Name     string
	IsEtf    bool
	Exchange string
	RssItems []*RSSItem `gorm:"many2many:security_rss_items;"`
}

func (MarketSecurity) TableName() string {
	return "market_securities"
}

type ItemTag struct {
	gorm.Model
	Name     string     `gorm:"unique_index"`
	RSSItems []*RSSItem `gorm:"many2many:item_tag_rss_items;"`
}

func (ItemTag) TableName() string {
	return "item_tags"
}

type ItemTagRSSItem struct {
	ItemTagID uint
	RSSItemID uint
}

func (ItemTagRSSItem) TableName() string {
	return "item_tag_rss_items"
}

type SecurityRssItem struct {
	SecurityID uint
	RSSItemID  uint
}

func (SecurityRssItem) TableName() string {
	return "security_rss_items"
}

func setupDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}

	// Auto migrate models
	err = db.AutoMigrate(&RSSFeed{}).Error
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&RSSItem{}).Error
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&MarketSecurity{}).Error
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&ItemTag{}).Error
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&ItemTagRSSItem{}).Error
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&SecurityRssItem{}).Error
	if err != nil {
		return nil, err
	}

	seedRSSFeeds(db)

	return db, nil
}

func seedRSSFeeds(db *gorm.DB) error {
	feeds := []RSSFeed{
		{
			Title:       "Technology Feed PR.com",
			Description: "This is the first feed",
			Link:        "https://www.pr.com/rss/news-by-category/170.xml",
		},
		{
			Title:       "Science Feed PR.com",
			Description: "The science feed",
			Link:        "https://www.pr.com/rss/news-by-category/141.xml",
		},
		{
			Title:       "Medical & Health PR.com",
			Description: "Medical and health news",
			Link:        "https://www.pr.com/rss/news-by-category/103.xml",
		},
		{
			Title:       "Semiconductor Industry PR.com",
			Description: "Semiconductor industry news",
			Link:        "https://www.pr.com/rss/news-by-category/188.xml",
		},
		// Add more feeds as needed
	}

	for _, feed := range feeds {
		var existingFeed RSSFeed
		if err := db.Where(&RSSFeed{Title: feed.Title}).First(&existingFeed).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				// Feed does not exist, create it
				if err := db.Create(&feed).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
