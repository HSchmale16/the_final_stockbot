package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	_ "net/http/pprof"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/app"
	"github.com/hschmale16/the_final_stockbot/internal/congress"
	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/travel"
	"github.com/hschmale16/the_final_stockbot/internal/votes"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"github.com/hschmale16/the_final_stockbot/pkg/utils"
	"github.com/robfig/cron/v3"
)

//go:generate npm run build

var reprocessId int = 0
var disableFetcherService = false
var disableWebServer = false
var loadCongressMembers = false
var congMemberFile = ""
var doSitemap = false
var scanLawText = false
var loadCclFile = ""
var doSenateLobbyingMain = false
var committeesFile = ""
var committeeMembershipsFile = ""

// New things will define a value in my switch case below for a script to run.
var script = ""
var file = ""

func init() {
	flag.StringVar(&file, "file", "", "some file to load")
	flag.StringVar(&script, "script", "", "Run a script")
	flag.IntVar(&reprocessId, "reprocess", 0, "Reprocess a specific item by ID")
	flag.BoolVar(&disableFetcherService, "disable-fetcher", false, "Disable the fetcher service")
	flag.BoolVar(&disableWebServer, "disable-web", false, "Disable the web server")
	flag.BoolVar(&loadCongressMembers, "load-congress-members", false, "Load congress members")
	flag.BoolVar(&scanLawText, "scan-law-text", false, "Scan law text")
	flag.BoolVar(&doSitemap, "sitemap", false, "Generate a sitemap")
	flag.StringVar(&congMemberFile, "congress-members-file", "", "The file to load congress members from")
	flag.StringVar(&loadCclFile, "load-ccl-file", "", "The file to load the CCL file from")
	flag.BoolVar(&doSenateLobbyingMain, "senate-lobbying-main", false, "Run the senate lobbying main")
	flag.StringVar(&committeesFile, "committees-file", "", "The file to load committees from")
	flag.StringVar(&committeeMembershipsFile, "committee-memberships-file", "", "The file to load committee memberships from")
}

func main() {
	flag.Parse()

	db, err := m.SetupDB()
	if err != nil {
		log.Fatal(err)
	}

	if script != "" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		// Run a random task
		//app.DoTagUpdates()
		//stocks.LoadDocuments(file)
		// stocks.ProcessBatchOfDocuments(db)
		fmt.Println("Target file is ", file)
		switch script {
		case "house-travel":
			utils.FindFileInZipUseCallback(file, func(rc io.ReadCloser) {
				travel.LoadHouseXml(rc, db)
			})
		case "senate-travel":
			utils.FindFileInZipUseCallback(file, func(rc io.ReadCloser) {
				travel.LoadSenateXml(rc, db)
			})
		case "house-votes":
			var scrape = map[int]int{
				2021: 449,
				2022: 549,
				2023: 724,
				2024: 400,
			}

			for year, maxRollCall := range scrape {
				for i := 1; i <= maxRollCall; i++ {
					url := fmt.Sprintf("https://clerk.house.gov/evs/%d/roll%03d.xml", year, i)
					votes.LoadHouseRollCallXml(url, db)
				}
			}
		}

		return
	}

	if doSenateLobbyingMain {
		senatelobbying.LoadFilings()
		return
	}

	if doSitemap {
		app.MakeSitemap()
		return
	}

	if reprocessId != 0 {
		app.FindUntaggedLaws()
		time.Sleep(10 * time.Second)
	}

	if committeesFile != "" {
		congress.LoadCongressCommittees(committeesFile)
		return
	}

	if committeeMembershipsFile != "" {
		congress.LoadCommitteeMemberships(committeeMembershipsFile)
		return
	}

	if loadCclFile != "" {
		x := fecwrangling.LoadLinkageZipFile(loadCclFile)

		for i := range x {
			//fmt.Println(i)
			db.Create(&i)
		}
		return
	}

	if scanLawText {
		app.LOAD_Members_Mods_2_RSS(db)
		return
	}

	if loadCongressMembers {
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
		// cron.AddFunc("@every 12", func() {
		// 	db.Exec("ANALYZE")
		// })
		//cron.AddFunc("@every 12h", app.FindUntaggedLaws)
		app.FindUntaggedLaws()
		triggerRssFetch()

		cron.Start()
	}

	if !disableWebServer {
		fmt.Println("Before go live do database maintenance")
		// On start up do database maintanence
		db.Exec("ANALYZE")

		fmt.Println("Starting up...")
		app.SetupServer()
		fmt.Println("The FUCK")
	}
	fmt.Println("Done!")
}
