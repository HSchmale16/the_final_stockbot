package fecwrangling // import "github.com/hschmale16/the_final_stockbot/internal/fec_wrangling"

import "fmt"

// https://www.fec.gov/campaign-finance-data/candidate-committee-linkage-file-description/
type CampaignCanidateLinkage struct {
	CandidateID          string
	ElectionYear         string
	FecElectionYear      string
	CommitteeID          string
	CommitteeType        string
	CommitteeDesignation string
	LinkageId            string
}

func LoadLinkageZipFile(zipFileName string) {
	fmt.Println("")
}
