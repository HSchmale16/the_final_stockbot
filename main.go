package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
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
	flag.Parse()
}

func doStartupTasks(db *gorm.DB) {
	if fetchFeedsFlag {
		fetchFeeds(db)
	}
	if downloadSymbolsFlag {
		downloadSecurities(db)
	}
}

func main() {
	db, err := setupDB()
	if err != nil {
		// Handle error
		log.Println("Error occurred:", err)
		return
	}
	defer db.Close()

	doStartupTasks(db)

	c := cron.New()
	c.AddFunc("@every 15m", func() {
		fetchFeeds(db)
	})

	c.AddFunc("@every 1m", func() {
		var firstItem RSSItem
		db.Where("id NOT IN (SELECT rss_item_id FROM item_tag_rss_items)").Order("pub_date desc").First(&firstItem)

		getRssItemTags(firstItem, db)
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
