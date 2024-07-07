package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/hschmale16/the_final_stockbot/internal/app"
	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"github.com/robfig/cron/v3"
)

var reprocessId int = 0
var disableFetcherService = false
var disableWebServer = false
var loadCongressMembers = false
var congMemberFile = ""
var doSitemap = false
var scanLawText = false
var loadCclFile = ""
var doSenateLobbyingMain = false

func init() {
	flag.IntVar(&reprocessId, "reprocess", 0, "Reprocess a specific item by ID")
	flag.BoolVar(&disableFetcherService, "disable-fetcher", false, "Disable the fetcher service")
	flag.BoolVar(&disableWebServer, "disable-web", false, "Disable the web server")
	flag.BoolVar(&loadCongressMembers, "load-congress-members", false, "Load congress members")
	flag.BoolVar(&scanLawText, "scan-law-text", false, "Scan law text")
	flag.BoolVar(&doSitemap, "sitemap", false, "Generate a sitemap")
	flag.StringVar(&congMemberFile, "congress-members-file", "", "The file to load congress members from")
	flag.StringVar(&loadCclFile, "load-ccl-file", "", "The file to load the CCL file from")
	flag.BoolVar(&doSenateLobbyingMain, "senate-lobbying-main", false, "Run the senate lobbying main")
}

func main() {
	flag.Parse()

	if doSenateLobbyingMain {
		senatelobbying.Main()
		return
	}

	if doSitemap {
		app.MakeSitemap()
		return
	}

	if loadCclFile != "" {
		db, err := app.SetupDB()
		if err != nil {
			log.Fatal(err)
		}
		x := fecwrangling.LoadLinkageZipFile(loadCclFile)

		for i := range x {
			//fmt.Println(i)
			db.Debug().Create(&i)
		}
		return
	}

	if scanLawText {
		db, err := app.SetupDB()
		if err != nil {
			log.Fatal(err)
		}
		app.LOAD_Members_Mods_2_RSS(db)
		return
	}

	if loadCongressMembers {
		db, err := app.SetupDB()
		if err != nil {
			log.Fatal(err)
		}
		app.LOAD_MEMBERS_JSON(db, congMemberFile)
		return
	}

	if !disableFetcherService {
		ch := make(app.LawRssItemChannel, 10)

		go app.RunFetcherService(ch)

		triggerRssFetch := func() {
			log.Println("Triggering RSS fetch")
			for _, rssLink := range app.RssLinks {
				go app.HandleLawRss(rssLink, ch)
			}
		}

		cron := cron.New()
		cron.AddFunc("@every 4h", triggerRssFetch)
		triggerRssFetch()

		cron.Start()
	}

	fmt.Println("Starting up...")

	if !disableWebServer {
		app.SetupServer()
	}
	fmt.Println("Done!")
}
