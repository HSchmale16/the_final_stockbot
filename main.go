package main

import (
	"flag"
	"log"
	_ "net/http/pprof"

	"github.com/robfig/cron/v3"
)

var reprocessId int = 0
var disableFetcherService = false
var disableWebServer = false

func init() {
	flag.IntVar(&reprocessId, "reprocess", 0, "Reprocess a specific item by ID")
	flag.BoolVar(&disableFetcherService, "disable-fetcher", false, "Disable the fetcher service")
	flag.BoolVar(&disableWebServer, "disable-web", false, "Disable the web server")
}

func main() {
	flag.Parse()

	if reprocessId > 0 {
		// Reprocess a specific item
		db, err := setupDB()
		if err != nil {
			panic(err)
		}

		var item GovtRssItem
		db.First(&item, reprocessId)

		var item2 GovtLawText
		db.First(&item2, "govt_rss_item_id = ?", item.ID)

		//ProcessLawTextForTags(item, db)
		processModsXML(item2.ModsXML)
	}

	if !disableFetcherService {

		ch := make(LawRssItemChannel, 10)

		go RunFetcherService(ch)

		triggerRssFetch := func() {
			log.Println("Triggering RSS fetch")
			for _, rssLink := range RssLinks {
				go handleLawRss(rssLink, ch)
			}
		}

		cron := cron.New()
		cron.AddFunc("@every 4h", triggerRssFetch)

		cron.Start()
	}

	// // Start the pprof server
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	if !disableWebServer {
		SetupServer()
	}

}
