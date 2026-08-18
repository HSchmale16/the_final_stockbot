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
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

type US_CongressLegislator = m.US_CongressLegislator

/**
 * https://github.com/unitedstates/congress-legislators?tab=readme-ov-file
 *
 * Legislators can be pulled from this github repo. The data is available from this json file:
 */

func fetchLegislatorJson(url string) []US_CongressLegislator {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

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

func GetCurrentLegislatorJson() []US_CongressLegislator {
	return fetchLegislatorJson("https://unitedstates.github.io/congress-legislators/legislators-current.json")
}

func GetHistoricalLegislatorJson() []US_CongressLegislator {
	return fetchLegislatorJson("https://unitedstates.github.io/congress-legislators/legislators-historical.json")
}

func MangleLegislatorsAndMerge(db *gorm.DB, memberData []US_CongressLegislator) (int, int) {
	// Use a database transaction to ensure that we don't have any partial data
	// And for speed
	tx := db.Begin()
	var newCount, updatedCount int
	for _, cong := range memberData {
		var existing DB_CongressMember
		err := tx.First(&existing, "bio_guide_id = ?", cong.Id.Bioguide).Error

		var myCongMember DB_CongressMember
		myCongMember.BioGuideId = cong.Id.Bioguide
		myCongMember.CongressMemberInfo = cong
		myCongMember.Name = cong.Name.Official
		if myCongMember.Name == "" {
			myCongMember.Name = fmt.Sprintf("%s %s", cong.Name.First, cong.Name.Last)
		}

		// set the is_active flag
		lastTerm := cong.Terms[len(cong.Terms)-1]
		termEnd, _ := time.Parse("2006-01-02", lastTerm.End)
		myCongMember.IsActive = time.Now().Before(termEnd)

		if err == gorm.ErrRecordNotFound {
			tx.Create(&myCongMember)
			newCount++
		} else {
			oldJSON, _ := json.Marshal(existing.CongressMemberInfo)
			newJSON, _ := json.Marshal(cong)
			if string(oldJSON) != string(newJSON) || existing.Name != myCongMember.Name || existing.IsActive != myCongMember.IsActive {
				myCongMember.CreatedAt = existing.CreatedAt
				tx.Save(&myCongMember)
				updatedCount++
			}
		}
	}
	tx.Commit()
	return newCount, updatedCount
}

func LOAD_MEMBERS_JSON(db *gorm.DB, file string) {
	var tCur []US_CongressLegislator
	if file == "" {
		// Fetch current and historical feeds concurrently
		type result struct {
			members []US_CongressLegislator
			err     error
		}
		ch := make(chan result, 2)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					ch <- result{err: fmt.Errorf("current feed panic: %v", r)}
				}
			}()
			ch <- result{members: GetCurrentLegislatorJson()}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ch <- result{err: fmt.Errorf("historical feed panic: %v", r)}
				}
			}()
			ch <- result{members: GetHistoricalLegislatorJson()}
		}()

		for i := 0; i < 2; i++ {
			r := <-ch
			if r.err != nil {
				log.Printf("WARNING: %v", r.err)
				continue
			}
			tCur = append(tCur, r.members...)
		}
		log.Printf("Loaded %d total legislators (current + historical)", len(tCur))
	} else {
		// open file by name
		jsonFile, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &tCur)
	}
	newCount, updatedCount := MangleLegislatorsAndMerge(db, tCur)
	m.LogCronJobRun(db, "update-congress-members", "success", newCount+updatedCount, fmt.Sprintf("Added %d new, updated %d members", newCount, updatedCount))
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

		modsData := m.ReadLawModsData(law.ModsXML)
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
