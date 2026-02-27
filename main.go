package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/pprof"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/app"
	"github.com/hschmale16/the_final_stockbot/internal/congress"
	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/travel"
	"github.com/hschmale16/the_final_stockbot/internal/votes"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"github.com/hschmale16/the_final_stockbot/pkg/utils"

	_ "net/http/pprof"

	_ "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"
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
		log.Fatal("Failed to connect: ", err)
	}

	if script != "" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	if !disableWebServer {
		pprof.Do(context.Background(), pprof.Labels("controller", "profiler"), func(c context.Context) {
			go runProfilerServer()
		})
		fmt.Println("Starting up...")
		app.SetupServer()
	}
	fmt.Println("Done!")
}

// Start pprof debug server on loopback only.
// Handlers are registered by the net/http/pprof blank import above.
// Nginx proxies /debug/pprof/ here after checking the secret header.
func runProfilerServer() {
	log.Println("pprof debug server listening on 127.0.0.1:6060")
	if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
		log.Println("pprof server error:", err)
	}
}
