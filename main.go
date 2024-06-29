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
var loadCongressMembers = false
var scanLawText = false

func init() {
	flag.IntVar(&reprocessId, "reprocess", 0, "Reprocess a specific item by ID")
	flag.BoolVar(&disableFetcherService, "disable-fetcher", false, "Disable the fetcher service")
	flag.BoolVar(&disableWebServer, "disable-web", false, "Disable the web server")
	flag.BoolVar(&loadCongressMembers, "load-congress-members", false, "Load congress members")
	flag.BoolVar(&scanLawText, "scan-law-text", false, "Scan law text")
}

func main() {
	flag.Parse()

	if scanLawText {
		db, err := setupDB()
		if err != nil {
			log.Fatal(err)
		}
		LOAD_Members_Mods_2_RSS(db)
		return
	}

	if loadCongressMembers {
		db, err := setupDB()
		if err != nil {
			log.Fatal(err)
		}
		CRON_LoadCongressMembers(db)
		return
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
		triggerRssFetch()

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
