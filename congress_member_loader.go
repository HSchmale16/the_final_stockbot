/**
 * This file is responsible for loading congress members into the database.
 * You'll note some fucky things in this file cause I'm sleepy while I'm building this.
 * And I just want to get the data loaded so I can do cool things with it.
 */

package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"
)

/**
 * https://github.com/unitedstates/congress-legislators?tab=readme-ov-file
 *
 * Legislators can be pulled from this github repo. The data is available from this json file:
* https://theunitedstates.io/congress-legislators/legislators-current.json
*/

func GetCurrentLegislatorJson() []US_CongressLegislator {
	resp, err := http.Get("https://theunitedstates.io/congress-legislators/legislators-current.json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the body of the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var congMembers []US_CongressLegislator
	err = json.Unmarshal(body, &congMembers)
	if err != nil {
		panic(err)
	}

	return congMembers
}

func MangleLegislatorsAndMerge(db *gorm.DB, memberData []US_CongressLegislator) {
	// Use a database transaction to ensure that we don't have any partial data
	// And for speed
	tx := db.Begin()
	for _, cong := range memberData {
		myCongMember := DB_CongressMember{
			BioGuideId:         cong.Id.Bioguide,
			CongressMemberInfo: cong,
		}

		tx.FirstOrCreate(&myCongMember, DB_CongressMember{BioGuideId: myCongMember.BioGuideId})
		myCongMember.CongressMemberInfo = cong
		myCongMember.Name = cong.Name.Official
		if myCongMember.Name == "" {
			myCongMember.Name = fmt.Sprintf("%s %s", cong.Name.First, cong.Name.Last)
		}

		fmt.Println(myCongMember.CongressMemberInfo.Terms)
		tx.Debug().Save(&myCongMember)
	}
	tx.Commit()
}

func LOAD_MEMBERS_JSON(db *gorm.DB, file string) {
	if file == "" {
		tCur := GetCurrentLegislatorJson()
		MangleLegislatorsAndMerge(db, tCur)
	} else {
		// open file by name
		jsonFile, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()
		// read file

		byteValue, _ := io.ReadAll(jsonFile)
		var tCur []US_CongressLegislator
		json.Unmarshal(byteValue, &tCur)
		MangleLegislatorsAndMerge(db, tCur)
	}
}

func LOAD_Members_Mods_2_RSS(db *gorm.DB) {
	// Step through law mods
	var lawTexts []GovtLawText
	page := 0
	for next := true; next; next = len(lawTexts) > 0 {
		db.Order("ID asc").Offset(page * 100).Limit(100).Find(&lawTexts)
		log.Print(len(lawTexts))
		for _, law := range lawTexts {
			lawData := ReadLawModsData(law.ModsXML)
			// Find the congress member
			for _, congMember := range lawData.CongressMembers {
				var dbCongMember DB_CongressMember
				db.Where("bio_guide_id = ?", congMember.BioGuideId).First(&dbCongMember)
				if dbCongMember.BioGuideId == "" {
					log.Printf("Could not find congress member %s\n", congMember)
					continue
				}
				// Create the association
				var dbCongMemberSponsored CongressMemberSponsored = CongressMemberSponsored{
					DB_CongressMemberBioGuideId: dbCongMember.BioGuideId,
					GovtRssItemId:               law.GovtRssItemId,
					Chamber:                     congMember.Chamber, // Slightly denormalized here. But it makes sense for the kind of questions we are asking and it can change over time. Trust the library of congress to get it right
					CongressNumber:              congMember.Congress,
					Role:                        congMember.Role,
				}
				result := db.Where(CongressMemberSponsored{
					DB_CongressMemberBioGuideId: dbCongMember.BioGuideId,
					GovtRssItemId:               law.GovtRssItemId,
				}).Assign(CongressMemberSponsored{
					Chamber:        dbCongMemberSponsored.Chamber,
					CongressNumber: dbCongMemberSponsored.CongressNumber,
					Role:           dbCongMemberSponsored.Role,
				}).FirstOrCreate(&dbCongMemberSponsored)

				fmt.Println(result.RowsAffected)
			}
		}
		page += 1
	}

}

type US_CongressLegislator struct {
	Id   CongIdentifiers `json:"id"`
	Name struct {
		First    string `json:"first"`
		Last     string `json:"last"`
		Official string `json:"official_full"`
	} `json:"name"`
	Bio struct {
		Birthday string `json:"birthday"`
		Gender   string `json:"gender"`
	} `json:"bio"`
	Terms []Terms `json:"terms"`
}

// Implement the scanner interface for US_CongressLegislators
func (l *US_CongressLegislator) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan US_CongressLegislators")
	}
	return json.Unmarshal(b, l)
}

// Implement the value interface for US_CongressLegislators
func (l US_CongressLegislator) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// See https://github.com/unitedstates/congress-legislators?tab=readme-ov-file
// For the data dictionary of each of these fields
type CongIdentifiers struct {
	Bioguide       string   `json:"bioguide"`
	Fec            []string `json:"fec"`
	Cspan          int      `json:"cspan"`
	Wikipedia      string   `json:"wikipedia"`
	HouseHistory   int      `json:"house_history"`
	Ballotpedia    string   `json:"ballotpedia"`
	Maplight       int      `json:"maplight"`
	Icpsr          int      `json:"icpsr"`
	Wikidata       string   `json:"wikidata"`
	GoogleEntityID string   `json:"google_entity_id"`
}

type Terms struct {
	Type  string `json:"type"`
	State string `json:"state"`
	Start string `json:"start"`
	End   string `json:"end"`
	Party string `json:"party"`
}

func (t Terms) IsSenator() bool {
	return t.Type == "sen"
}
