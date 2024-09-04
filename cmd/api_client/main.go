package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/congress"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	congressgov "github.com/hschmale16/the_final_stockbot/pkg/congress_gov"
	"gorm.io/datatypes"
)

func main() {
	db, err := m.SetupDB()
	if err != nil {
		panic(err)
	}

	client := congressgov.NewClient(os.Getenv("CONGRESS_GOV_API_TOKEN"))

	maxItems := 100000
	for i := 0; i < maxItems; i += 250 {
		data, err := client.GetBillsFromCongress(118, i)
		if err != nil {
			panic(err)
		}

		maxItems = data.Pagination.Count

		for i, bill := range data.Bills {
			fmt.Println("Working on entry", i, "of", len(data.Bills))
			dbBill := congress.Bill{
				CongressNumber: bill.Congress,
				BillNumber:     bill.Number,
				BillType:       bill.Type,
				Title:          bill.Title,
			}

			db.Create(&dbBill)

			// grab json blob of bill to store
			billData, err := client.GetBillDetails(bill.Congress, bill.Number, bill.Type)
			if err != nil {
				panic(err)
			}
			dbBill.JsonBlob = datatypes.JSON(billData)
			dbErr := db.Save(&dbBill)
			if dbErr.Error != nil {
				panic(dbErr.Error)
			}

			// Get the actions
			actions, err := client.GetBillActions(bill.Congress, bill.Number, bill.Type)
			if err != nil {
				panic(err)
			}

			for _, action := range actions.Actions {
				actionTimeStr := action.ActionDate
				actionTime, err := time.Parse(time.DateOnly, actionTimeStr)
				if err != nil {
					panic(err)
				}

				committeeId := sql.NullString{}
				if len(action.Committees) > 0 {
					stripped, _ := strings.CutSuffix(strings.ToUpper(action.Committees[0].SystemCode), "00")
					committeeId = sql.NullString{
						String: stripped,
						Valid:  true,
					}
				}

				dbAction := congress.BillAction{
					ActionTime:       actionTime,
					ActionCode:       action.ActionCode,
					ActionDate:       action.ActionDate,
					SourceSystemCode: action.SourceSystem.SourceSystemCode,
					Type:             action.Type,
					Text:             action.ActionText,
					BillID:           dbBill.ID,
					CommitteeId:      committeeId,
				}

				dbErr := db.Debug().Create(&dbAction)
				if dbErr.Error != nil {
					panic(dbErr.Error)
				}
			}

			// Get the cosponsors
			cosponsors, err := client.GetBillCosponsors(0, bill.Congress, bill.Number, bill.Type)
			if err != nil {
				panic(err)
			}

			if cosponsors.Pagination.Count > 250 {
				// Do a 2nd request and join htem
				cosponsors2, err := client.GetBillCosponsors(250, bill.Congress, bill.Number, bill.Type)
				if err != nil {
					panic(err)
				}

				cosponsors.Cosponsors = append(cosponsors.Cosponsors, cosponsors2.Cosponsors...)
			}

			for _, cosponsor := range cosponsors.Cosponsors {
				dbCosponsor := congress.BillCosponsor{
					BillID:            dbBill.ID,
					MemberId:          cosponsor.BioGuideId,
					OriginalCosponsor: cosponsor.OriginalCosponsor,
					SponsorshipDate:   cosponsor.SponsorshipDate,
				}

				err := db.Create(&dbCosponsor)
				if err.Error != nil {
					panic(err.Error)
				}
			}
		}
	}
}
