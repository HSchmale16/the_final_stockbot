package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/robfig/cron/v3"
)

var reprocessId int = 0
var disableFetcherService = false
var disableWebServer = false
var loadCongressMembers = false
var congMemberFile = ""
var doSitemap = false
var scanLawText = false

func init() {
	flag.IntVar(&reprocessId, "reprocess", 0, "Reprocess a specific item by ID")
	flag.BoolVar(&disableFetcherService, "disable-fetcher", false, "Disable the fetcher service")
	flag.BoolVar(&disableWebServer, "disable-web", false, "Disable the web server")
	flag.BoolVar(&loadCongressMembers, "load-congress-members", false, "Load congress members")
	flag.BoolVar(&scanLawText, "scan-law-text", false, "Scan law text")
	flag.BoolVar(&doSitemap, "sitemap", false, "Generate a sitemap")
	flag.StringVar(&congMemberFile, "congress-members-file", "", "The file to load congress members from")
}

func main() {
	flag.Parse()

	if doSitemap {
		MakeSitemap()
		return
	}

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
		LOAD_MEMBERS_JSON(db, congMemberFile)
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

	if !disableWebServer {
		SetupServer()
	}
	fmt.Println("Done!")
}
