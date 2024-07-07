package senatelobbying

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

func Main() {
	pprofFile, pprofErr := os.Create("cpu.pb.gz")
	if pprofErr != nil {
		log.Fatal(pprofErr)
	}
	pprof.StartCPUProfile(pprofFile)
	defer pprof.StopCPUProfile()
	// Dummy Main to Do Things

	url := GetContributionListUrl(ContributionListingFilterParams{
		FilingYear: "2019",
	})

	res, err := SendRequest(url)

	if err != nil {
		fmt.Printf("Error: %s\nBody: %s", err.Error(), string(res))
		panic(err)
	}

	var list []ContributionListing = make([]ContributionListing, 0, 50000)
	var response ContributionListResponse

	err = json.Unmarshal(res, &response)
	if err != nil {
		panic(err)
	}
	list = append(list, response.Results...)
	WriteToDatabase(response.Results)

	// while the response.Next is present we want to keep making requests
	// and appending the results to the list
	errCount := 0
	for response.Next != "" {
		res, err = SendRequest(response.Next)
		if err != nil {
			if strings.HasPrefix(err.Error(), "retry") {
				// If we get a rate limit error, we should wait and retry
				retryAfter, err := strconv.Atoi(strings.TrimPrefix(err.Error(), "retry "))
				if err != nil {
					panic(err)
				}
				WriteArray(list)
				fmt.Println("Sleeping for ", retryAfter, " seconds")
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			} else {
				// Otherwise WE PANIC
				fmt.Println(response.Next, err.Error())
				errCount++
				if errCount > 5 {
					panic(err)
				}
				continue
			}
		} else {
			// Reset the Error Count Because it's 5 errors in a row we die on.
			errCount = 0
		}

		err = json.Unmarshal(res, &response)
		if err != nil {
			fmt.Println(response.Next, string(res))
			panic(err)
		}
		list = append(list, response.Results...)
		WriteToDatabase(response.Results)

		fmt.Println(len(list), "of", response.Count, response.Next)

		if len(list) > response.Count {
			break
		}
	}

	WriteArray(list)
}

func WriteToDatabase(x []ContributionListing) {
	// Dummy Write to Database
	fmt.Println("Writing to Database")

	db, err := sql.Open("sqlite3", "file:contribution_list.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, item := range x {
		xjson, err := json.Marshal(item)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("INSERT INTO contributions(uuid, json_item) VALUES (?, ?) ON CONFLICT(uuid) DO UPDATE SET json_item = excluded.json_item", item.FilingUuid, xjson)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func WriteArray(x []ContributionListing) {
	outfile, err := os.Create("contribution_list.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	err = enc.Encode(x)
	if err != nil {
		panic(err)
	}
}
