package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
var doTravelBackgroundProcessing = false
var runHearingFetcher = false
var backfillCongress int = 0
var processHearings = false

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
	flag.BoolVar(&doTravelBackgroundProcessing, "travel-background-processing", false, "Run the travel background processing")
	flag.BoolVar(&runHearingFetcher, "run-hearing-fetcher", false, "Run the hearing fetcher service")
	flag.IntVar(&backfillCongress, "backfill-congress", 0, "Backfill hearings for a specific congress")
	flag.BoolVar(&processHearings, "process-hearings", false, "Process hearing attendance from transcripts")
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
		case "load-voteview":
			congress := 119
			if len(flag.Args()) > 0 {
				fmt.Sscanf(flag.Arg(0), "%d", &congress)
			} else if len(os.Args) > 3 {
				fmt.Sscanf(os.Args[3], "%d", &congress)
			} else if len(os.Args) > 2 {
				fmt.Sscanf(os.Args[2], "%d", &congress)
			}
			err := votes.LoadVotesFromVoteview(db, congress)
			if err != nil {
				log.Fatalf("Voteview load failed: %v", err)
			}
		case "house-travel":
			var createdCount int
			utils.FindFileInZipUseCallback(file, func(rc io.ReadCloser) {
				createdCount = travel.LoadHouseXml(rc, db)
			})
			m.LogCronJobRun(db, "house-travel", "success", createdCount, fmt.Sprintf("Successfully imported %d new House travel disclosures", createdCount))
		case "senate-travel":
			var createdCount int
			utils.FindFileInZipUseCallback(file, func(rc io.ReadCloser) {
				createdCount = travel.LoadSenateXml(rc, db)
			})
			m.LogCronJobRun(db, "senate-travel", "success", createdCount, fmt.Sprintf("Successfully imported %d new Senate travel disclosures", createdCount))
		case "house-votes":
			var createdCount int
			currentYear := time.Now().Year()
			for year := 2021; year <= currentYear; year++ {
				// Query database to find the maximum roll call number already imported for this year
				start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

				var maxRollCall int
				db.Model(&votes.Vote{}).
					Where("chamber = ? AND action_at BETWEEN ? AND ?", "U.S. House of Representatives", start, end).
					Select("COALESCE(MAX(roll_call_num), 0)").
					Row().Scan(&maxRollCall)

				log.Printf("Resuming House votes for year %d from roll %d\n", year, maxRollCall+1)

				consecutiveErrors := 0
				for i := maxRollCall + 1; ; i++ {
					url := fmt.Sprintf("https://clerk.house.gov/evs/%d/roll%03d.xml", year, i)
					err := votes.LoadHouseRollCallXml(url, db)
					if err != nil {
						log.Printf("Stopped House scraping for year %d at roll %d: %v\n", year, i, err)
						consecutiveErrors++
						if consecutiveErrors >= 3 {
							break
						}
					} else {
						createdCount++
						consecutiveErrors = 0
					}
				}
			}
			m.LogCronJobRun(db, "house-votes", "success", createdCount, fmt.Sprintf("Successfully processed %d House roll calls", createdCount))
		case "senate-votes":
			var createdCount int
			currentYear := time.Now().Year()

			for year := 2021; year <= currentYear; year++ {
				congress, session := getCongressSession(year)
				if congress == 0 {
					continue
				}

				// Find the highest roll call number for this congress and session in the database
				var maxRollCall int
				db.Model(&votes.Vote{}).
					Where("chamber = ? AND congress_num = ? AND session = ?", "Senate", congress, fmt.Sprintf("%d", session)).
					Select("COALESCE(MAX(roll_call_num), 0)").
					Row().Scan(&maxRollCall)

				log.Printf("Resuming Senate votes for Congress %d, Session %d (Year %d) from roll %d\n", congress, session, year, maxRollCall+1)

				consecutiveErrors := 0
				for i := maxRollCall + 1; ; i++ {
					url := fmt.Sprintf("https://www.senate.gov/legislative/LIS/roll_call_votes/vote%d%d/vote_%d_%d_%05d.xml", congress, session, congress, session, i)
					err := votes.LoadSenateRollCallXml(url, db)
					if err != nil {
						log.Printf("Stopped Senate scraping for Congress %d, Session %d at vote %d: %v\n", congress, session, i, err)
						consecutiveErrors++
						if consecutiveErrors >= 3 {
							break
						}
					} else {
						createdCount++
						consecutiveErrors = 0
					}
				}
			}
			m.LogCronJobRun(db, "senate-votes", "success", createdCount, fmt.Sprintf("Successfully processed %d Senate roll calls", createdCount))
		case "backfill-last-vote-date":
			log.Println("Backfilling last vote date for all congress members...")
			// Using COALESCE to keep existing dates if they have no votes, or set to subquery max.
			// This updates the denormalized column on the congress_member table.
			err := db.Exec(`
				UPDATE congress_member
				SET last_vote_date = (
					SELECT MAX(votes.action_at)
					FROM vote_records
					INNER JOIN votes ON votes.id = vote_records.vote_id
					WHERE vote_records.member_id = congress_member.bio_guide_id
				)
			`).Error
			if err != nil {
				log.Fatalf("Failed to backfill last vote date: %v\n", err)
			}
			log.Println("Backfill completed successfully.")
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

	pprof.Do(context.Background(), pprof.Labels("controller", "profiler"), func(c context.Context) {
		go runProfilerServer()
	})

	if doTravelBackgroundProcessing {
		pprof.Do(context.Background(), pprof.Labels("controller", "travel background processing"), func(c context.Context) {
			processor := travel.NewBackgroundProcessor(db)
			processor.ProcessDisclosuresInBackground()
		})
		return
	}

	if runHearingFetcher {
		app.RunHearingFetcherService()
		return
	}

	if backfillCongress != 0 {
		app.RunHearingBackfill(backfillCongress)
		return
	}

	if processHearings {
		app.ProcessHearingAttendance(db)
		return
	}

	if !disableWebServer {
		pprof.Do(context.Background(), pprof.Labels("controller", "app setup"), func(c context.Context) {
			localApp := app.SetupServer()
			pprof.Do(context.Background(), pprof.Labels("controller", "app listen"), func(c context.Context) {
				err = localApp.Listen(":8080")
				if err != nil {
					log.Fatal("Something failed during app listen", err)
				}
			})
		})

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

func getCongressSession(year int) (int, int) {
	if year < 1789 {
		return 0, 0
	}
	yearsSince1789 := year - 1789
	congress := 1 + (yearsSince1789 / 2)
	session := 1 + (yearsSince1789 % 2)
	return congress, session
}
