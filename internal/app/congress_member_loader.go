/**
 * This file is responsible for loading congress members into the database.
 * You'll note some fucky things in this file cause I'm sleepy while I'm building this.
 * And I just want to get the data loaded so I can do cool things with it.
 */

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"

	. "github.com/hschmale16/the_final_stockbot/internal/m"
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
		tx.Save(&myCongMember)
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
	type foo struct {
		A string
		B uint
	}
	numCongress := make(map[foo]bool)

	rows, err := db.Model(&GovtLawText{}).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var law GovtLawText
		db.ScanRows(rows, &law)

		db.Model(&law).Association("GovtRssItem").Find(&law.GovtRssItem)

		modsData := ReadLawModsData(law.ModsXML)
		law.GovtRssItem.Metadata = modsData
		db.Save(&law.GovtRssItem)

		for _, member := range modsData.CongressMembers {
			numCongress[foo{member.BioGuideId, law.GovtRssItemId}] = true
		}

		ScanLawSponsors(modsData, law.GovtRssItem, db)
		if law.GovtRssItem.Metadata.IsAppropriation {
			CreateTagsOnItem([]string{"Appropriation"}, law.GovtRssItem, 0, db)
		}

	}
	var x int64
	db.Model(&CongressMemberSponsored{}).Count(&x)
	fmt.Println(len(numCongress), x)
}
