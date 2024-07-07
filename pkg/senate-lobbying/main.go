package senatelobbying

import (
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
		FilingYear: "2021",
	})

	res, err := SendRequest(url)

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error(), string(res))
		panic(err)
	}

	var list []ContributionListing = make([]ContributionListing, 0, 50000)
	var response ContributionListResponse

	err = json.Unmarshal(res, &response)
	if err != nil {
		panic(err)
	}

	// while the response.next is present we want to keep making requests
	// and appending the results to the list
	for response.Next != "" {
		res, err = SendRequest(response.Next)
		if err != nil {
			if strings.HasPrefix(err.Error(), "retry") {
				// If we get a rate limit error, we should wait and retry
				retryAfter, err := strconv.Atoi(strings.TrimPrefix(err.Error(), "retry "))
				if err != nil {
					panic(err)
				}
				fmt.Println("Sleeping for ", retryAfter, " seconds")
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			} else {
				// Otherwise WE PANIC
				panic(err)
			}
		}

		list = append(list, response.Results...)
		err = json.Unmarshal(res, &response)
		if err != nil {
			fmt.Println(response.Next, string(res))
			panic(err)
		}

		fmt.Println(len(list), "of", response.Count, response.Next)

		if len(list) > response.Count {
			break
		}
	}

	outfile, err := os.Create("contribution_list.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	err = enc.Encode(list)
	if err != nil {
		panic(err)
	}
}
