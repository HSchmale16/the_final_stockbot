package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/vertexai/genai"
	"github.com/hschmale16/the_final_stockbot/internal/app"
	"github.com/hschmale16/the_final_stockbot/internal/congress"
	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/travel"
	"github.com/hschmale16/the_final_stockbot/internal/votes"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"github.com/hschmale16/the_final_stockbot/pkg/utils"
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

// Downloads the target pdf then uploads it to GCS
func downloadUrlUploadToGCS(url string) (string, error) {

	// Download the file
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer res.Body.Close()

	// Upload the file to GCS
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	// split the url as the first / is the bucket name
	arr := strings.SplitN(url, "/", 4)
	objName := arr[3]

	fmt.Println("Doing things")
	// Upload the file to GCS using the first / as the separator for the path
	bucket := client.Bucket("dirtycongress-test")
	obj := bucket.Object(objName)
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, res.Body); err != nil {
		return "", fmt.Errorf("failed to upload file to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return objName, nil
}

func TestAiStuff(filePath string) error {
	// gs://dirtycongress-test/500028402.pdf

	ctx := context.Background()

	projectID := "dirtycongress-88971"
	modelName := "gemini-1.5-flash-001"
	location := "us-central1"

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)

	part := genai.FileData{
		MIMEType: "application/pdf",
		FileURI:  filePath,
	}

	// res, err := model.GenerateContent(ctx, part, genai.Text(`Please identify the amount of money spent on travel in this trip disclosure.
	// Please break it down by lodging, travel and meal expenses.
	// Also identify if there was a spouse or significant other on this trip.`))
	res, err := model.GenerateContent(ctx, part, genai.Text(`
	Please identify the companies involved in this trip and the topics discussed. Keep them to broad industries and specifically named companies.
	`))
	if err != nil {
		return fmt.Errorf("unable to generate contents: %w", err)
	}

	fmt.Println(res.UsageMetadata.CandidatesTokenCount, "output tokens used")
	fmt.Println(res.UsageMetadata.PromptTokenCount, "prompt tokens used")

	if len(res.Candidates) == 0 ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return errors.New("empty response from model")
	}

	fmt.Printf("generated response: %s\n", res.Candidates[0].Content.Parts[0])
	return nil
}

func main() {
	flag.Parse()

	db, err := m.SetupDB()
	if err != nil {
		log.Fatal("Failed to connect: ", err)
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
		case "ai-stuff":
			var bullshit []travel.DB_TravelDisclosure
			db.Model(&travel.DB_TravelDisclosure{}).Where("length(doc_url) > 0").Limit(10).Find(&bullshit)

			path, err := downloadUrlUploadToGCS(bullshit[0].DocURL)
			if err != nil {
				log.Fatalf("Failed to download and upload to GCS: %v", err)
			}
			fmt.Println(path)

			// err = TestAiStuff()
			// if err != nil {
			// 	log.Fatalf("Failed to run AI stuff: %v", err)
			// }
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

	// if !disableFetcherService {
	// 	ch := make(app.LawRssItemChannel, 10)

	// 	go app.RunFetcherService(ch)

	// 	triggerRssFetch := func() {
	// 		log.Println("Triggering RSS fetch")
	// 		for _, rssLink := range app.RssLinks {
	// 			go app.HandleLawRss(rssLink, ch)
	// 		}
	// 	}

	// 	cron := cron.New()
	// 	cron.AddFunc("@every 4h", triggerRssFetch)
	// 	// cron.AddFunc("@every 12", func() {
	// 	// 	db.Exec("ANALYZE")
	// 	// })
	// 	//cron.AddFunc("@every 12h", app.FindUntaggedLaws)
	// 	app.FindUntaggedLaws()
	// 	triggerRssFetch()

	// 	cron.Start()
	// }

	if !disableWebServer {
		// fmt.Println("Before go live do database maintenance")
		// // On start up do database maintanence
		// db.Exec("ANALYZE")

		fmt.Println("Starting up...")
		app.SetupServer()
		fmt.Println("The FUCK")
	}
	fmt.Println("Done!")
}
