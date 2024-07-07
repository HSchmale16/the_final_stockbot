package senatelobbying

const BASE_URL = "https://lda.senate.gov/api/v1/"

type FilingListResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []FilingListing `json:"results"`
}

type FilingListing struct {
	FilingType string `json:"filing_type"`
}
