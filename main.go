package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
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

		ProcessLawTextForTags(item, db)
	}

	if !disableFetcherService {
		go RunFetcherService()
	}

	// Start the pprof server
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if !disableWebServer {
		SetupServer()
	}

}
