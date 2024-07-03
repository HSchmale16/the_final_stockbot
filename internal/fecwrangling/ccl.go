package fecwrangling // import "github.com/hschmale16/the_final_stockbot/internal/fec_wrangling"

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"log"
)

/*
 * https://www.fec.gov/campaign-finance-data/candidate-committee-linkage-file-description/
 */
type CampaignCanidateLinkage struct {
	CandidateID          string
	ElectionYear         string
	FecElectionYear      string
	CommitteeID          string
	CommitteeType        string
	CommitteeDesignation string
	LinkageId            string `gorm:"primaryKey"`
}

func (CampaignCanidateLinkage) TableName() string {
	return "campaign_canidate_linkage"
}

type CampaignCanidateLinkageChannel chan CampaignCanidateLinkage

/** Helps to load an FEC Campaign Canidate Linkage File
 */
func LoadLinkageZipFile(zipFileName string) CampaignCanidateLinkageChannel {
	ch := make(CampaignCanidateLinkageChannel, 20)

	go func() {
		defer close(ch)

		// Open the zip file
		r, err := zip.OpenReader(zipFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		// Find the ccl.txt file
		var cclFile *zip.File
		for _, f := range r.File {
			if f.Name == "ccl.txt" {
				cclFile = f
				break
			}
		}

		if cclFile == nil {
			log.Fatal("ccl.txt not found in zip file")
		}

		// Open the ccl.txt file
		rc, err := cclFile.Open()
		if err != nil {
			log.Fatal(err)
		}

		// Read the file as | separated values
		reader := csv.NewReader(rc)
		reader.Comma = '|'

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break // End of file
			}

			if err != nil {
				log.Fatal(err)
			}

			linkage := CampaignCanidateLinkage{
				CandidateID:          record[0],
				ElectionYear:         record[1],
				FecElectionYear:      record[2],
				CommitteeID:          record[3],
				CommitteeType:        record[4],
				CommitteeDesignation: record[5],
				LinkageId:            record[6],
			}

			ch <- linkage
		}

		// Parse the CSV into a struct
		// Send the struct to the channel

		defer rc.Close()

	}()
	return ch
}
