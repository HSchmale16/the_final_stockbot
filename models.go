package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GovtRssItem struct {
	gorm.Model
	DescriptiveMetaUrl string
	FullTextUrl        string
	Title              string
	Link               string `gorm:"unique_index"`
	PubDate            time.Time
	ProcessedOn        time.Time
}

func (GovtRssItem) TableName() string {
	return "govt_rss_item"
}

/**
 * GovtLawText is the full text of a law item fetched from the FullTextUrl
 */
type GovtLawText struct {
	gorm.Model

	GovtRssItemId uint
	GovtRssItem   GovtRssItem
	Text          string
}

func (GovtLawText) TableName() string {
	return "govt_law_text"
}

/** Tag is a simple tag for categorizing items */
type Tag struct {
	gorm.Model
	Name string `gorm:"unique"`
}

func (Tag) TableName() string {
	return "tag"
}

/** GovtRssItemTag is a many-to-many relationship between GovtRssItem and Tag */
type GovtRssItemTag struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
	ID         uint

	GovtRssItemId uint `gorm:"index:,unique,composite:myname"`
	TagId         uint `gorm:"index:,unique,composite:myname"`
	Metadata      string

	GovtRssItem GovtRssItem
	Tag         Tag
}

// Add a compound unique index on GovtRssItemId and TagId
func (GovtRssItemTag) TableName() string {
	return "govt_rss_item_tag"
}

/**
 * Sets up the stupid database
 */
func setupDB() (*gorm.DB, error) {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
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
	if err := db.AutoMigrate(&GovtRssItem{}, &GovtLawText{}, &Tag{}, &GovtRssItemTag{}); err != nil {
		return nil, err
	}

	return db, nil
}
