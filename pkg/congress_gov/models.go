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
}

type BillActions struct {
	Actions []struct {
		ActionCode   string `json:"actionCode"`
		ActionDate   string `json:"actionDate"`
		ActionText   string `json:"text"`
		SourceSystem struct {
			SourceSystemCode int    `json:"code"`
			SourceSystemName string `json:"name"`
		}
		Type string `json:"type"`
	} `json:"actions"`
}
