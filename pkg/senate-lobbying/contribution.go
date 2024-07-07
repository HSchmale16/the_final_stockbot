package senatelobbying

import "net/url"

type ContributionListResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next"`
	Previous string                `json:"previous"`
	Results  []ContributionListing `json:"results"`
}

type ContributionListing struct {
	Url                       string `json:"url"`
	FilingUuid                string `json:"filing_uuid"`
	FilingType                string `json:"filing_type"`
	FilingTypeDisplay         string `json:"filing_type_display"`
	FilingYear                int    `json:"filing_year"`
	FilingPeriod              string `json:"filing_period"`
	FilingPeriodDisplay       string `json:"filing_period_display"`
	FilingDocumentUrl         string `json:"filing_document_url"`
	FilingDocumentContentType string `json:"filing_document_content_type"`
	FilerType                 string `json:"filer_type"`
	FilerTypeDisplay          string `json:"filer_type_display"`
	DtPosted                  string `json:"dt_posted"`
	// There's a contact_name field but it seems to always be null
	Comments       string        `json:"comments"`
	Address1       string        `json:"address_1"`
	Address2       string        `json:"address_2"`
	City           string        `json:"city"`
	State          string        `json:"state"`
	StateDisplay   string        `json:"state_display"`
	Zip            string        `json:"zip"`
	Country        string        `json:"country"`
	CountryDisplay string        `json:"country_display"`
	Registraint    S_Registraint `json:"registrant"`

	Lobbyist S_Lobbyist `json:"lobbyist"`

	NoContributions   bool               `json:"no_contributions"`
	Pacs              []string           `json:"pacs"`
	ContributionItems []ContributionItem `json:"contribution_items"`
}

type S_Lobbyist struct {
	Id            int    `json:"id"`
	Prefix        string `json:"prefix"`
	PrefixDisplay string `json:"prefix_display"`
	FirstName     string `json:"first_name"`
	Nickname      string `json:"nickname"`
	MiddleName    string `json:"middle_name"`
	LastName      string `json:"last_name"`
	Suffix        string `json:"suffix"`
	SuffixDisplay string `json:"suffix_display"`
}

type S_Registraint struct {
	Id                int    `json:"id"`
	Url               string `json:"url"`
	HouseRegistrantId int    `json:"house_registrant_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Address1          string `json:"address_1"`
	Address2          string `json:"address_2"`
	Address3          string `json:"address_3"`
	Address4          string `json:"address_4"`
	City              string `json:"city"`
	State             string `json:"state"`
	StateDisplay      string `json:"state_display"`
	Zip               string `json:"zip"`
	Country           string `json:"country"`
	CountryDisplay    string `json:"country_display"`
	PpbCountry        string `json:"ppb_country"`
	PpbCountryDisplay string `json:"ppb_country_display"`
	ContactName       string `json:"contact_name"`
	ContactPhone      string `json:"contact_telephone"`
	DTUpdated         string `json:"dt_updated"`
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
