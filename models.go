package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

type RSSFeed struct {
	gorm.Model
	Title       string
	Description string
	Link        string
	PubDate     time.Time
}

func (RSSFeed) TableName() string {
	return "rss_feeds"
}

