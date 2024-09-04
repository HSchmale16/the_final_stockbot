package congressgov

type LatestBillActions struct {
	Bills []struct {
		Congress     int `json:"congress"`
		LatestAction struct {
			ActionDate string `json:"actionDate"`
			Text       string `json:"text"`
		} `json:"latestAction"`
		Number                  string `json:"number"`
		OriginChamber           string `json:"originChamber"`
		OriginChamberCode       string `json:"originChamberCode"`
		Title                   string `json:"title"`
		Type                    string `json:"type"`
		UpdateDate              string `json:"updateDate"`
		UpdateDateIncludingText string `json:"updateDateIncludingText"`
		URL                     string `json:"url"`
	} `json:"bills"`
	Pagination Pagination `json:"pagination"`
}

type BillActions struct {
	Actions []BillAction `json:"actions"`
}

type Pagination struct {
	Count   int    `json:"count"`
	NextUrl string `json:"nextUrl"`
}

type CosponsorsResponse struct {
	Cosponsors []struct {
		BioGuideId        string `json:"bioguideid"`
		OriginalCosponsor bool   `json:"isOriginalCosponsor"`
		SponsorshipDate   string `json:"sponsorshipDate"`
	} `json:"cosponsors"`
	Pagination Pagination `json:"pagination"`
}

type BillAction struct {
	ActionCode   string `json:"actionCode"`
	ActionDate   string `json:"actionDate"`
	ActionTime   string `json:"actionTime"`
	ActionText   string `json:"text"`
	SourceSystem struct {
		SourceSystemCode int    `json:"code"`
		SourceSystemName string `json:"name"`
	}
	RecordedVotes []struct {
		Chamber       string `json:"chamber"`
		Congress      int    `json:"congress"`
		Date          string `json:"date"`
		RollNumber    int    `json:"rollNumber"`
		SessionNumber int    `json:"sessionNumber"`
		Url           string `json:"url"`
	} `json:"recordedVotes"`
	CalendarNumber struct {
		Calendar string `json:"calendar"`
		Number   string `json:"number"`
	} `json:"calendarNumber"`
	Committees []struct {
		Name string `json:"name"`
		// Links to the thomas_id
		SystemCode string `json:"systemCode"`
	} `json:"committees"`
	Type string `json:"type"`
}
