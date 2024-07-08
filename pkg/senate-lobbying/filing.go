package senatelobbying

type FilingListResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []FilingListing `json:"results"`
}

type FilingListing struct {
	Url                   string        `json:"url"`
	FilingUuid            string        `json:"filing_uuid"`
	FilingType            string        `json:"filing_type"`
	FilingTypeDisplay     string        `json:"filing_type_display"`
	FilingYear            int           `json:"filing_year"`
	FilingPeriod          string        `json:"filing_period"`
	FilingPeriodDisplay   string        `json:"filing_period_display"`
	FilingDocumentUrl     string        `json:"filing_document_url"`
	FilingDocumentType    string        `json:"filing_document_type"`
	Income                string        `json:"income"`
	Expenses              string        `json:"expenses"`
	ExpensesMethod        string        `json:"expenses_method"`
	ExpensesMethodDisplay string        `json:"expenses_method_display"`
	PostedByname          string        `json:"posted_by_name"`
	DtPosted              string        `json:"dt_posted"`
	TerminationDate       string        `json:"termination_date"`
	RegistratiantCountry  string        `json:"registrant_country"`
	Registraint           S_Registraint `json:"registrant"`
	Client                S_Client      `json:"client"`

	// Big old lists
	LobbyingActivities      []LobbyingActivities   `json:"lobbying_activities"`
	ConvictionDisclosures   []ConvictionDisclosure `json:"conviction_disclosures"`
	ForeignEntities         []ForeignEntity        `json:"foreign_entities"`
	AffiliatedOrganizations []S_Registraint        `json:"affiliated_organizations"`
}

type S_Client struct {
	Id                     int    `json:"id"`
	Url                    string `json:"url"`
	ClientId               int    `json:"client_id"`
	Name                   string `json:"name"`
	GeneralDescription     string `json:"general_description"`
	ClientGovernmentEntity bool   `json:"client_government_entity"`
	ClientSelfSelect       bool   `json:"client_self_select"`
	State                  string `json:"state"`
	Country                string `json:"country"`
	EffectiveDate          string `json:"effective_date"`
}

type LobbyingActivities struct {
	GeneralIssueCode        string `json:"general_issue_code"`
	GeneralIssueCodeDisplay string `json:"general_issue_code_display"`
	Description             string `json:"description"`
	ForeignEntityIssues     string `json:"foreign_entity_issues"`
	Lobbyists               []FilingLobbyist
}

type FilingLobbyist struct {
	Lobbyist        S_Lobbyist `json:"lobbyist"`
	CoveredPosition string     `json:"covered_position"`
	New             bool       `json:"new"`
}

type GovernmentEntity struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ConvictionDisclosure struct {
	Lobbyist    S_Lobbyist `json:"lobbyist"`
	Date        string     `json:"date"`
	Description string     `json:"description"`
}

type ForeignEntity struct {
	Name                string `json:"name"`
	Contribution        string `json:"contribution"`
	OwnershipPercentage string `json:"ownership_percentage"`
	Address             string `json:"address"`
	City                string `json:"city"`
	State               string `json:"state"`
	StateDisplay        string `json:"state_display"`
	Country             string `json:"country"`
	CountryDisplay      string `json:"country_display"`
	PpbCity             string `json:"ppb_city"`
	PpbState            string `json:"ppb_state"`
	PpbStateDisplay     string `json:"ppb_state_display"`
	PpbCountry          string `json:"ppb_country"`
	PpbCountryDisplay   string `json:"ppb_country_display"`
}

type FilingListingFilterParams struct {
	FilingYear string
}

func GetFilingListUrl(params FilingListingFilterParams) string {
	return "https://lda.senate.gov/api/v1/filings/?filing_year=" + params.FilingYear
}
