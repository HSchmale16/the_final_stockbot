package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"gorm.io/gorm"
)

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

var (
	fetchFeedsFlag      bool
	downloadSymbolsFlag bool
)

func init() {
	flag.BoolVar(&fetchFeedsFlag, "fetchFeeds", false, "Enable fetching feeds task")
	flag.BoolVar(&downloadSymbolsFlag, "downloadSymbols", false, "Enable downloading securities task")
}

func doStartupTasks(db *gorm.DB) {
	flag.Parse()

	if fetchFeedsFlag {
		fetchFeeds(db)
	}
	if downloadSymbolsFlag {
		downloadSecurities(db)
	}
}

func pickAnRssItemToScan(db *gorm.DB) RSSItem {
	modelID := 3
	var firstItem RSSItem
	db.Debug().Where("id NOT IN (SELECT rss_item_id FROM item_tag_rss_items WHERE model_id = ?)", modelID).Order("pub_date desc").First(&firstItem)
	return firstItem
}

func main() {
	db, err := setupDB()
	if err != nil {
		// Handle error
		log.Println("Error occurred:", err)
		return
	}

	doStartupTasks(db)

	c := cron.New()
	c.AddFunc("@every 6m", func() {
		fetchFeeds(db)
	})

	c.AddFunc("@every 38s", func() {
		// Count RSS Items that have not been tagged and print it
		var count int64
		db.Model(&RSSItem{}).Where("id NOT IN (SELECT rss_item_id FROM item_tag_rss_items)").Count(&count)
		fmt.Println("RSS Items to tag:", count)

		if count != 0 {
			firstItem := pickAnRssItemToScan(db)

			getRssItemTags(firstItem, db)
		}
	})

	c.AddFunc("0 0 * * *", func() {
		downloadSecurities(db)
	})

	log.Print("Started feed reader cron.")
	c.Start()

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	router.GET("/checkLoadedStories", GetLoadedArticlesStatus)

	router.Run(":8080")
}
