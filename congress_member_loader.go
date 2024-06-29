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
	"io"
	"net/http"

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
	for _, cong := range memberData {
		myCongMember := DB_CongressMember{
			BioGuideId:         cong.Id.Bioguide,
			CongressMemberInfo: cong,
		}

		db.FirstOrCreate(&myCongMember, DB_CongressMember{BioGuideId: myCongMember.BioGuideId})
		myCongMember.CongressMemberInfo = cong
		myCongMember.Name = cong.Name.Official
		db.Debug().Save(&myCongMember)
	}
}

func CRON_LoadCongressMembers(db *gorm.DB) {
	tCur := GetCurrentLegislatorJson()

	MangleLegislatorsAndMerge(db, tCur)
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
