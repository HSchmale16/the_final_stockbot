package main

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

type LLMModel struct {
	gorm.Model
	ModelName string
}

type ItemTagRSSItem struct {
	gorm.Model
	ItemTagID uint
	RSSItemID uint
	ModelID   uint
	LLM       LLMModel `gorm:"foreignkey:ModelID"`
	ItemTag   ItemTag  `gorm:"foreignkey:ItemTagID"`
	RSSItem   RSSItem  `gorm:"foreignkey:RSSItemID"`
}

func (ItemTagRSSItem) TableName() string {
	return "item_tag_rss_items"
}

type SecurityRssItem struct {
	SecurityID uint
	RSSItemID  uint
	ModelID    uint
}

func (SecurityRssItem) TableName() string {
	return "security_rss_items"
}

func setupDB() (*gorm.DB, error) {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info, // Log level Info will output everything
		},
	)

	// Globally mode
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate models
	if err := db.AutoMigrate(&RSSFeed{}, &RSSItem{}, &MarketSecurity{}, &SecurityRssItem{}, &LLMModel{}, &ItemTag{}, &ItemTagRSSItem{}); err != nil {
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
		{
			Title:       "Deals Reuters",
			Description: "Reuters deals news",
			Link:        "https://www.reutersagency.com/feed/?best-topics=deals&post_type=best",
		},
		{
			Title:       "Market Impact Reuters",
			Description: "Reuters market impact news",
			Link:        "https://www.reutersagency.com/feed/?best-customer-impacts=market-impact&post_type=best",
		},
		{
			Title:       "Reuters Health",
			Description: "Reuters health news",
			Link:        "https://www.reutersagency.com/feed/?best-topics=health&post_type=best",
		},
		{
			Title:       "Aljazeera",
			Description: "Middle eastern news source",
			Link:        "https://www.aljazeera.com/xml/rss/all.xml",
		},
		{
			Title:       "Cipher Brief",
			Description: "Some random blog about Global Security",
			Link:        "https://www.thecipherbrief.com/feed",
		},
		{
			Title:       "United Nations Top Stories",
			Description: "Who gives a fuck",
			Link:        "https://news.un.org/feed/subscribe/en/news/all/rss.xml",
		},
		{
			Title:       "US State Dept Direct Line to American Business",
			Description: "US State Department News",
			Link:        "https://www.state.gov/rss-feed/direct-line-to-american-business/feed/",
		},
		{
			Title:       "US State Dept Europe and Eurasisa",
			Description: "Who cares",
			Link:        "https://www.state.gov/rss-feed/europe-and-eurasia/feed/",
		},
		{
			Title:       "Economic, Energy, Agricultural and Trade Issues &#8211; United States Department of State",
			Description: "Economic, Energy, Agricultural and Trade Issues &#8211; United States Department of State",
			Link:        "https://www.state.gov/rss-feed/economic-energy-agricultural-and-trade-issues/feed/",
		},
		// Add more feeds as needed
	}

	for _, feed := range feeds {
		var existingFeed RSSFeed
		if err := db.Where(&RSSFeed{Title: feed.Title}).First(&existingFeed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Feed does not exist, create it
				if err := db.Create(&feed).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	models := []LLMModel{
		{
			ModelName: "phi3",
		},
		{
			ModelName: "gemma:2.6b",
		},
	}
	for _, model := range models {
		var existingModel LLMModel
		if err := db.Where(&LLMModel{ModelName: model.ModelName}).First(&existingModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Model does not exist, create it
				if err := db.Create(&model).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Model already exists, update it
			existingModel.ModelName = model.ModelName
			if err := db.Save(&existingModel).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
