package votes

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func LoadVotesFromVoteview(db *gorm.DB, congressNum int) error {
	log.Printf("Starting Voteview bulk load for Congress %d...\n", congressNum)

	// 1. Download member mappings (icpsr -> bioguide_id)
	log.Println("Downloading UCLA Voteview member mappings...")
	membersUrl := "https://voteview.com/static/data/out/members/HSall_members.csv"
	resp, err := http.Get(membersUrl)
	if err != nil {
		return fmt.Errorf("failed to download member mappings: %v", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	icpsrIdx := -1
	bioguideIdx := -1
	for i, h := range header {
		if h == "icpsr" {
			icpsrIdx = i
		} else if h == "bioguide_id" {
			bioguideIdx = i
		}
	}

	if icpsrIdx == -1 || bioguideIdx == -1 {
		return fmt.Errorf("invalid member mapping CSV header")
	}

	icpsrToBioguide := make(map[string]string)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		icpsr := row[icpsrIdx]
		bioguide := strings.TrimSpace(row[bioguideIdx])
		if icpsr != "" && bioguide != "" {
			icpsrToBioguide[icpsr] = bioguide
		}
	}
	log.Printf("Loaded %d member ID mappings.\n", len(icpsrToBioguide))

	// 2. Download rollcalls for this congress
	log.Printf("Downloading UCLA Voteview rollcalls for Congress %d...\n", congressNum)
	rollcallsUrl := fmt.Sprintf("https://voteview.com/static/data/out/rollcalls/HS%d_rollcalls.csv", congressNum)
	resp, err = http.Get(rollcallsUrl)
	if err != nil {
		return fmt.Errorf("failed to download rollcalls: %v", err)
	}
	defer resp.Body.Close()

	reader = csv.NewReader(resp.Body)
	header, err = reader.Read()
	if err != nil {
		return err
	}

	colMap := make(map[string]int)
	for i, h := range header {
		colMap[h] = i
	}

	var voteList []Vote
	voteMap := make(map[string]Vote) // key: chamber + "|" + rollnumber

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rawChamber := row[colMap["chamber"]]
		chamber := "Senate"
		if strings.EqualFold(rawChamber, "House") {
			chamber = "U.S. House of Representatives"
		}

		rollNum, _ := strconv.Atoi(row[colMap["rollnumber"]])
		dateStr := row[colMap["date"]]
		actionAt, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// fallback to noon
			actionAt = time.Now()
		} else {
			// Set to noon UTC to avoid timezone issues
			actionAt = time.Date(actionAt.Year(), actionAt.Month(), actionAt.Day(), 12, 0, 0, 0, time.UTC)
		}

		sessionVal := row[colMap["session"]]
		sessionStr := sessionVal
		if chamber == "U.S. House of Representatives" {
			if sessionVal == "1" {
				sessionStr = "1st"
			} else if sessionVal == "2" {
				sessionStr = "2nd"
			}
		}

		billNum := row[colMap["bill_number"]]
		voteResult := row[colMap["vote_result"]]
		voteDesc := row[colMap["vote_desc"]]
		voteQuestion := row[colMap["vote_question"]]

		fakeUrl := fmt.Sprintf("https://voteview.com/rollcall?congress=%d&chamber=%s&roll=%d", congressNum, rawChamber, rollNum)

		v := Vote{
			RollCallNum:      rollNum,
			CongressNum:      congressNum,
			Session:          sessionStr,
			Chamber:          chamber,
			ActionAt:         actionAt,
			VoteType:         voteQuestion,
			LegisName:        billNum,
			VoteResult:       voteResult,
			VoteDesc:         voteDesc,
			Url:              fakeUrl,
		}

		voteList = append(voteList, v)
		voteMap[rawChamber+"|"+strconv.Itoa(rollNum)] = v
	}

	log.Printf("Importing %d votes (roll calls) into database...\n", len(voteList))
	for _, v := range voteList {
		err = db.Where(Vote{
			RollCallNum: v.RollCallNum,
			CongressNum: v.CongressNum,
			Session:     v.Session,
			Chamber:     v.Chamber,
		}).Attrs(v).FirstOrCreate(&v).Error
		if err != nil {
			return fmt.Errorf("failed to insert vote: %v", err)
		}
		// update map with DB-assigned ID
		rawChamber := "Senate"
		if v.Chamber == "U.S. House of Representatives" {
			rawChamber = "House"
		}
		voteMap[rawChamber+"|"+strconv.Itoa(v.RollCallNum)] = v
	}

	// 3. Download individual votes
	log.Printf("Downloading UCLA Voteview individual votes for Congress %d...\n", congressNum)
	votesUrl := fmt.Sprintf("https://voteview.com/static/data/out/votes/HS%d_votes.csv", congressNum)
	resp, err = http.Get(votesUrl)
	if err != nil {
		return fmt.Errorf("failed to download individual votes: %v", err)
	}
	defer resp.Body.Close()

	reader = csv.NewReader(resp.Body)
	header, err = reader.Read()
	if err != nil {
		return err
	}

	for i, h := range header {
		colMap[h] = i
	}

	// Load existing member IDs to filter missing members
	var existingIds []string
	if err = db.Model(&m.DB_CongressMember{}).Pluck("bio_guide_id", &existingIds).Error; err != nil {
		return err
	}
	existingMap := make(map[string]bool)
	for _, id := range existingIds {
		existingMap[id] = true
	}

	var records []VoteRecord
	castCodes := map[string]string{
		"1": "Yea",
		"2": "Yea", // Paired Yea
		"3": "Yea", // Announced Yea
		"4": "Nay", // Announced Nay
		"5": "Nay", // Paired Nay
		"6": "Nay",
		"7": "Present",
		"8": "Present",
		"9": "Not Voting",
	}

	countSkipped := 0
	countAdded := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rawChamber := row[colMap["chamber"]]
		rollStr := row[colMap["rollnumber"]]
		icpsr := row[colMap["icpsr"]]
		castCode := row[colMap["cast_code"]]

		bioguide, ok := icpsrToBioguide[icpsr]
		if !ok {
			countSkipped++
			continue
		}

		if !existingMap[bioguide] {
			countSkipped++
			continue
		}

		voteObj, ok := voteMap[rawChamber+"|"+rollStr]
		if !ok {
			continue
		}

		status := castCodes[castCode]
		if status == "" {
			status = "Not Voting"
		}

		records = append(records, VoteRecord{
			VoteID:     voteObj.ID,
			MemberId:   bioguide,
			VoteStatus: status,
		})
		countAdded++
	}

	log.Printf("Upserting %d individual vote records in batches (skipped %d)...", len(records), countSkipped)

	// Upsert on the composite unique constraint (vote_id, member_id).
	// This is idempotent and avoids primary-key sequence sync issues from prior partial runs.
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vote_id"}, {Name: "member_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"vote_status"}),
	}).CreateInBatches(&records, 500).Error
	if err != nil {
		return fmt.Errorf("failed to upsert vote records: %v", err)
	}

	log.Println("Voteview bulk load completed successfully.")
	return nil
}
