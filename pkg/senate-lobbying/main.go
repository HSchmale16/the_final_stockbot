package senatelobbying

import (
	"encoding/json"
	"fmt"
)

func Main() {
	// Dummy Main to Do Things

	url := GetContributionListUrl(ContributionListingFilterParams{
		FilingYear: "2021",
	})

	res, err := SendRequest(url)

	if err != nil {
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
		list = append(list, response.Results...)
		res, err = SendRequest(response.Next)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(res, &response)
		if err != nil {
			panic(err)
		}

		fmt.Println(len(list), "of", response.Count)
	}

	fmt.Println(list)

}
