package senatelobbying

import "net/url"

type ContributionListResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next"`
	Previous string                `json:"previous"`
	Results  []ContributionListing `json:"results"`
}

type ContributionListing struct {
	NoContributions   bool               `json:"no_contributions"`
	ContributionItems []ContributionItem `json:"contribution_items"`
}

type ContributionItem struct {
	Amount           string `json:"amount"`
	PayeeName        string `json:"payee_name"`
	HonreeName       string `json:"honoree_name"`
	RecipientName    string `json:"recipient_name"`
	ContributionType string `json:"contribution_type"`
	Date             string `json:"date"`
}

type ContributionListingFilterParams struct {
	FilingYear string
}

func GetContributionListUrl(p ContributionListingFilterParams) string {
	baseUrl, err := url.Parse(BASE_URL + "contributions/")
	if err != nil {
		panic(err)
	}

	params := url.Values{}
	if p.FilingYear != "" {
		params.Add("filing_year", p.FilingYear)
	}

	baseUrl.RawQuery = params.Encode()

	return baseUrl.String()
}
