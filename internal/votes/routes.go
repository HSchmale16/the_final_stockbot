package votes

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	vote := app.Group("/htmx/votes")
	vote.Get("/member/:memberId", GetVotesForMember)
	vote.Get("/:voteId", GetVote)

	app.Get("/attendance", RedirectToCurrentYearAttendance)
	app.Get("/attendance/:year", GetAttendanceYearPage)
	app.Get("/htmx/attendance/:year", GetHtmxAttendanceScoreboard)
}

func RedirectToCurrentYearAttendance(c *fiber.Ctx) error {
	currentYear := time.Now().Year()
	qs := string(c.Request().URI().QueryString())
	if qs != "" {
		return c.Redirect(fmt.Sprintf("/attendance/%d?%s", currentYear, qs), fiber.StatusFound)
	}
	return c.Redirect(fmt.Sprintf("/attendance/%d", currentYear), fiber.StatusFound)
}

func GetVote(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	voteId := c.Params("voteId")

	var vote Vote
	db.Debug().Preload("VoteRecords").Preload("VoteRecords.Member").First(&vote, "id = ?", voteId)

	// Bin them by vote_status

	var voteStatus = map[string][]VoteRecord{}
	for _, record := range vote.VoteRecords {
		voteStatus[record.VoteStatus] = append(voteStatus[record.VoteStatus], record)
	}

	return c.Render("vote_table", fiber.Map{
		"Title":      "Vote",
		"Vote":       vote,
		"VoteStatus": voteStatus,
	})
}

func GetVotesForMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	memberId := c.Params("memberId")

	// Get member details to check if they exist / display name
	var member m.DB_CongressMember
	if err := db.First(&member, "bio_guide_id = ?", memberId).Error; err != nil {
		return c.Status(404).SendString("Member not found")
	}

	// Fetch all records for this member preloading Vote to compute attendance and streaks
	var records []VoteRecord
	db.Preload("Vote").
		Joins("INNER JOIN votes ON votes.id = vote_records.vote_id").
		Where("vote_records.member_id = ?", memberId).
		Order("votes.action_at DESC").
		Find(&records)

	totalVotes := len(records)
	missedVotes := 0
	yeaVotes := 0
	nayVotes := 0
	presentVotes := 0

	// Calculate streaks and grouping
	currentStreak := 0
	streakBroken := false
	currentMissedStreak := 0
	longestMissedStreak := 0

	// Visual timeline of last 20 votes
	type TimelineItem struct {
		ID         uint
		Voted      bool
		VoteStatus string
		VoteDesc   string
		DateStr    string
		Result     string
		Chamber    string
		RollCall   int
		ColorClass string
	}
	var timeline []TimelineItem

	// Vote list table
	type VoteTableItem struct {
		ID            uint
		DateStr       string
		Chamber       string
		RollCall      int
		Question      string
		VoteCast      string
		Result        string
		LegisName     string
		AmendmentInfo string
	}
	var recentVotes []VoteTableItem

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	for _, r := range records {
		statusClean := strings.TrimSpace(strings.ToLower(r.VoteStatus))
		isMissed := statusClean == "not voting" || statusClean == "absent"

		if isMissed {
			missedVotes++
			streakBroken = true
			currentMissedStreak++
			if currentMissedStreak > longestMissedStreak {
				longestMissedStreak = currentMissedStreak
			}
		} else {
			currentMissedStreak = 0
			if !streakBroken {
				currentStreak++
			}
			if strings.Contains(statusClean, "yea") || strings.Contains(statusClean, "aye") {
				yeaVotes++
			} else if strings.Contains(statusClean, "nay") || strings.Contains(statusClean, "no") {
				nayVotes++
			} else if strings.Contains(statusClean, "present") {
				presentVotes++
			}
		}

		// Keep timeline up to 20 items
		if len(timeline) < 20 {
			color := "bg-slate-500"
			if isMissed {
				color = "bg-red-500"
			}
			timeline = append(timeline, TimelineItem{
				ID:         r.Vote.ID,
				Voted:      !isMissed,
				VoteStatus: r.VoteStatus,
				VoteDesc:   r.Vote.VoteDesc,
				DateStr:    r.Vote.ActionAt.Format("Jan 02, 2006"),
				Result:     r.Vote.VoteResult,
				Chamber:    r.Vote.Chamber,
				RollCall:   r.Vote.RollCallNum,
				ColorClass: color,
			})
		}

		// Table items for votes in the last 30 days
		if r.Vote.ActionAt.After(thirtyDaysAgo) {
			amdtInfo := ""
			if r.Vote.AmmendmentAuthor != "" {
				amdtInfo = fmt.Sprintf("Amdt %d (%s)", r.Vote.AmmendmentNum, r.Vote.AmmendmentAuthor)
			}
			recentVotes = append(recentVotes, VoteTableItem{
				ID:            r.Vote.ID,
				DateStr:       r.Vote.ActionAt.Format("Jan 02, 2006"),
				Chamber:       r.Vote.Chamber,
				RollCall:      r.Vote.RollCallNum,
				Question:      r.Vote.VoteDesc,
				VoteCast:      r.VoteStatus,
				Result:        r.Vote.VoteResult,
				LegisName:     r.Vote.LegisName,
				AmendmentInfo: amdtInfo,
			})
		}
	}

	// Reverse timeline so oldest votes are on the left, newest on the right
	for i, j := 0, len(timeline)-1; i < j; i, j = i+1, j-1 {
		timeline[i], timeline[j] = timeline[j], timeline[i]
	}

	attendanceRate := 0.0
	if totalVotes > 0 {
		attendanceRate = float64(totalVotes-missedVotes) / float64(totalVotes) * 100
	}

	return c.Render("votes_members", fiber.Map{
		"Member":              member,
		"TotalVotes":          totalVotes,
		"MissedVotes":         missedVotes,
		"AttendedVotes":       totalVotes - missedVotes,
		"AttendanceRate":      fmt.Sprintf("%.1f", attendanceRate),
		"CurrentStreak":       currentStreak,
		"LongestMissedStreak": longestMissedStreak,
		"Timeline":            timeline,
		"RecentVotes":         recentVotes,
		"YeaCount":            yeaVotes,
		"NayCount":            nayVotes,
		"PresentCount":        presentVotes,
	})
}

type ScoreboardItem struct {
	BioGuideId          string
	Name                string
	Party               string
	State               string
	Chamber             string
	IsSenator           bool
	TotalVotes          int
	MissedVotes         int
	AttendanceRate      float64
	MaxMissedStreak     int
	RecentMissedVotes   int
	RecentEligibleVotes int
	Age                 int
	YearsServed         int
}

func computeAttendanceScores(db *gorm.DB, year int, chamberFilter, tenureFilter, ageFilter string) ([]ScoreboardItem, map[string]interface{}, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	yearStart := fmt.Sprintf("%d-01-01", year)
	yearEnd := fmt.Sprintf("%d-12-31", year)

	var allMembers []m.DB_CongressMember
	err := db.Select("bio_guide_id, name, is_active, congress_member_info").
		Where("EXISTS (SELECT 1 FROM jsonb_array_elements(congress_member_info->'terms') AS t WHERE (t->>'end') >= ? AND (t->>'start') <= ?)", yearStart, yearEnd).
		Find(&allMembers).Error
	if err != nil {
		return nil, nil, err
	}

	type VoteQueryRow struct {
		VoteId     uint
		MemberId   string
		VoteStatus string
		ActionAt   time.Time
		OrderIdx   int
	}
	type VoteItem struct {
		ID       uint
		ActionAt time.Time
	}
	var yearVotes []VoteItem
	err = db.Table("votes").
		Select("id, action_at").
		Where("action_at BETWEEN ? AND ?", start, end).
		Order("action_at ASC, id ASC").
		Scan(&yearVotes).Error
	if err != nil {
		return nil, nil, err
	}

	if len(yearVotes) == 0 {
		return nil, map[string]interface{}{
			"OverallAvg":  "0.0",
			"SenateAvg":   "0.0",
			"HouseAvg":    "0.0",
			"TopLeader":   nil,
			"WorstLeader": nil,
		}, nil
	}

	voteDateMap := make(map[uint]time.Time, len(yearVotes))
	voteOrderMap := make(map[uint]int, len(yearVotes))
	voteIDs := make([]uint, len(yearVotes))
	for i, v := range yearVotes {
		voteIDs[i] = v.ID
		voteDateMap[v.ID] = v.ActionAt
		voteOrderMap[v.ID] = i
	}

	type RecordRow struct {
		VoteId     uint
		MemberId   string
		VoteStatus string
	}
	var records []RecordRow
	err = db.Table("vote_records").
		Select("vote_id, member_id, vote_status").
		Where("vote_id IN ?", voteIDs).
		Scan(&records).Error
	if err != nil {
		return nil, nil, err
	}

	memberRecords := make(map[string][]VoteQueryRow, len(allMembers))
	for _, r := range records {
		memberRecords[r.MemberId] = append(memberRecords[r.MemberId], VoteQueryRow{
			VoteId:     r.VoteId,
			MemberId:   r.MemberId,
			VoteStatus: r.VoteStatus,
			ActionAt:   voteDateMap[r.VoteId],
			OrderIdx:   voteOrderMap[r.VoteId],
		})
	}

	var allScores []ScoreboardItem
	now := time.Now()

	for _, member := range allMembers {
		term, inOffice := member.CongressMemberInfo.TermForYear(year)
		if !inOffice {
			continue
		}

		state := term.State
		// Exclude non-voting delegates and territory members (American Samoa, DC, Guam, Northern Mariana Islands, Puerto Rico, Virgin Islands)
		if state == "AS" || state == "DC" || state == "GU" || state == "MP" || state == "PR" || state == "VI" {
			continue
		}

		birthday, err := time.Parse("2006-01-02", member.CongressMemberInfo.Bio.Birthday)
		age := 0
		if err == nil {
			age = now.Year() - birthday.Year()
			if now.YearDay() < birthday.YearDay() {
				age--
			}
		}

		yearsServed := member.YearsServed()
		mRecords := memberRecords[member.BioGuideId]
		var eligibleRecords []VoteQueryRow
		for _, r := range mRecords {
			if member.CongressMemberInfo.WasInOfficeOn(r.ActionAt) {
				eligibleRecords = append(eligibleRecords, r)
			}
		}

		if len(eligibleRecords) == 0 {
			continue
		}

		sort.Slice(eligibleRecords, func(i, j int) bool {
			return eligibleRecords[i].OrderIdx < eligibleRecords[j].OrderIdx
		})

		totalVotes := len(eligibleRecords)
		missedVotes := 0
		currentMissedStreak := 0
		maxMissedStreak := 0

		for _, r := range eligibleRecords {
			statusClean := strings.TrimSpace(strings.ToLower(r.VoteStatus))
			isMissed := statusClean == "not voting" || statusClean == "absent"

			if isMissed {
				missedVotes++
				currentMissedStreak++
				if currentMissedStreak > maxMissedStreak {
					maxMissedStreak = currentMissedStreak
				}
			} else {
				currentMissedStreak = 0
			}
		}

		recentMissed := 0
		recentEligible := 0
		thirtyDaysAgo := now.AddDate(0, 0, -30)
		for _, r := range eligibleRecords {
			if r.ActionAt.After(thirtyDaysAgo) {
				recentEligible++
				statusClean := strings.TrimSpace(strings.ToLower(r.VoteStatus))
				if statusClean == "not voting" || statusClean == "absent" {
					recentMissed++
				}
			}
		}

		attendanceRate := 0.0
		if totalVotes > 0 {
			attendanceRate = float64(totalVotes-missedVotes) / float64(totalVotes) * 100.0
			attendanceRate = float64(int(attendanceRate*10+0.5)) / 10.0
		}

		chamber := "Senate"
		if term.Type == "rep" {
			chamber = "House"
		}

		partyName := "Independent"
		partyClean := strings.ToUpper(term.Party)
		if strings.HasPrefix(partyClean, "D") {
			partyName = "Democrat"
		} else if strings.HasPrefix(partyClean, "R") {
			partyName = "Republican"
		}

		allScores = append(allScores, ScoreboardItem{
			BioGuideId:          member.BioGuideId,
			Name:                member.Name,
			Party:               partyName,
			State:               state,
			Chamber:             chamber,
			IsSenator:           chamber == "Senate",
			TotalVotes:          totalVotes,
			MissedVotes:         missedVotes,
			AttendanceRate:      attendanceRate,
			MaxMissedStreak:     maxMissedStreak,
			RecentMissedVotes:   recentMissed,
			RecentEligibleVotes: recentEligible,
			Age:                 age,
			YearsServed:         yearsServed,
		})
	}

	var senateSum, houseSum, overallSum float64
	var senateCount, houseCount, overallCount int
	var topLeader, worstLeader ScoreboardItem
	hasTop, hasWorst := false, false

	for _, s := range allScores {
		if s.TotalVotes < 15 {
			continue
		}

		overallSum += s.AttendanceRate
		overallCount++

		if s.IsSenator {
			senateSum += s.AttendanceRate
			senateCount++
		} else {
			houseSum += s.AttendanceRate
			houseCount++
		}

		if !hasTop || s.AttendanceRate > topLeader.AttendanceRate {
			topLeader = s
			hasTop = true
		} else if s.AttendanceRate == topLeader.AttendanceRate && s.TotalVotes > topLeader.TotalVotes {
			topLeader = s
		}

		if !hasWorst || s.MaxMissedStreak > worstLeader.MaxMissedStreak {
			worstLeader = s
			hasWorst = true
		}
	}

	overallAvg := 0.0
	if overallCount > 0 {
		overallAvg = float64(int(overallSum/float64(overallCount)*10+0.5)) / 10.0
	}
	senateAvg := 0.0
	if senateCount > 0 {
		senateAvg = float64(int(senateSum/float64(senateCount)*10+0.5)) / 10.0
	}
	houseAvg := 0.0
	if houseCount > 0 {
		houseAvg = float64(int(houseSum/float64(houseCount)*10+0.5)) / 10.0
	}

	var topLeaderVal interface{} = nil
	if hasTop {
		topLeaderVal = topLeader
	}
	var worstLeaderVal interface{} = nil
	if hasWorst {
		worstLeaderVal = worstLeader
	}

	metadata := map[string]interface{}{
		"OverallAvg":  fmt.Sprintf("%.1f", overallAvg),
		"SenateAvg":   fmt.Sprintf("%.1f", senateAvg),
		"HouseAvg":    fmt.Sprintf("%.1f", houseAvg),
		"TopLeader":   topLeaderVal,
		"WorstLeader": worstLeaderVal,
		"TotalCount":  len(allScores),
	}

	var filtered []ScoreboardItem
	for _, s := range allScores {
		if chamberFilter != "All" {
			if chamberFilter == "Senate" && !s.IsSenator {
				continue
			}
			if chamberFilter == "House" && s.IsSenator {
				continue
			}
		}

		if tenureFilter != "All" {
			if tenureFilter == "newcomer" && s.YearsServed >= 4 {
				continue
			}
			if tenureFilter == "established" && (s.YearsServed < 4 || s.YearsServed > 12) {
				continue
			}
			if tenureFilter == "veteran" && s.YearsServed <= 12 {
				continue
			}
		}

		if ageFilter != "All" {
			if ageFilter == "under50" && s.Age >= 50 {
				continue
			}
			if ageFilter == "50to65" && (s.Age < 50 || s.Age > 65) {
				continue
			}
			if ageFilter == "over65" && s.Age <= 65 {
				continue
			}
		}

		filtered = append(filtered, s)
	}

	return filtered, metadata, nil
}

func GetAttendanceYearPage(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	yearStr := c.Params("year")
	var year int
	if yearStr == "" {
		year = time.Now().Year()
	} else {
		fmt.Sscanf(yearStr, "%d", &year)
	}
	currentYear := time.Now().Year()
	if year < 2021 {
		year = currentYear
	}

	chamber := c.Query("chamber", "All")
	tenure := c.Query("tenure", "All")
	age := c.Query("age", "All")

	filtered, metadata, err := computeAttendanceScores(db, year, chamber, tenure, age)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Database error: %v", err))
	}

	topAttendees := make([]ScoreboardItem, len(filtered))
	copy(topAttendees, filtered)
	sort.Slice(topAttendees, func(i, j int) bool {
		if topAttendees[i].AttendanceRate == topAttendees[j].AttendanceRate {
			return topAttendees[i].TotalVotes > topAttendees[j].TotalVotes
		}
		return topAttendees[i].AttendanceRate > topAttendees[j].AttendanceRate
	})
	limitTop := 5
	if len(topAttendees) < limitTop {
		limitTop = len(topAttendees)
	}
	top5 := topAttendees[:limitTop]
	worstStreaks := make([]ScoreboardItem, len(filtered))
	copy(worstStreaks, filtered)
	sort.Slice(worstStreaks, func(i, j int) bool {
		if worstStreaks[i].MaxMissedStreak == worstStreaks[j].MaxMissedStreak {
			return worstStreaks[i].AttendanceRate < worstStreaks[j].AttendanceRate
		}
		return worstStreaks[i].MaxMissedStreak > worstStreaks[j].MaxMissedStreak
	})
	limitWorst := 5
	if len(worstStreaks) < limitWorst {
		limitWorst = len(worstStreaks)
	}
	worst5 := worstStreaks[:limitWorst]

	limitRecent := 20
	var recentlyMissing []ScoreboardItem
	if year == currentYear {
		for _, s := range filtered {
			if s.RecentMissedVotes > 0 {
				recentlyMissing = append(recentlyMissing, s)
			}
		}
		sort.Slice(recentlyMissing, func(i, j int) bool {
			if recentlyMissing[i].RecentMissedVotes == recentlyMissing[j].RecentMissedVotes {
				return recentlyMissing[i].AttendanceRate < recentlyMissing[j].AttendanceRate
			}
			return recentlyMissing[i].RecentMissedVotes > recentlyMissing[j].RecentMissedVotes
		})
		if len(recentlyMissing) > limitRecent {
			recentlyMissing = recentlyMissing[:limitRecent]
		}
	}

	// Sort full roster for the by-year table
	allScoresTable := make([]ScoreboardItem, len(filtered))
	copy(allScoresTable, filtered)
	sort.Slice(allScoresTable, func(i, j int) bool {
		if allScoresTable[i].AttendanceRate == allScoresTable[j].AttendanceRate {
			return allScoresTable[i].TotalVotes < allScoresTable[j].TotalVotes
		}
		return allScoresTable[i].AttendanceRate < allScoresTable[j].AttendanceRate
	})

	availableYears := make([]string, 0, currentYear-2020)
	for y := currentYear; y >= 2021; y-- {
		availableYears = append(availableYears, strconv.Itoa(y))
	}

	selectedYearStr := year

	chambers := []map[string]string{
		{"Value": "All", "Label": "All"},
		{"Value": "House", "Label": "House"},
		{"Value": "Senate", "Label": "Senate"},
	}
	tenures := []map[string]string{
		{"Value": "All", "Label": "All"},
		{"Value": "newcomer", "Label": "< 4 Yrs"},
		{"Value": "established", "Label": "4-12 Yrs"},
		{"Value": "veteran", "Label": "> 12 Yrs"},
	}
	ages := []map[string]string{
		{"Value": "All", "Label": "All"},
		{"Value": "under50", "Label": "< 50"},
		{"Value": "50to65", "Label": "50-65"},
		{"Value": "over65", "Label": "> 65"},
	}

	hasFilters := chamber != "All" || tenure != "All" || age != "All"

	voteFreq := computeYearVoteFrequency(db, year, chamber)

	bindMap := fiber.Map{
		"Title":          fmt.Sprintf("Attendance Scoreboard (%d)", year),
		"SelectedYear":   selectedYearStr,
		"IsCurrentYear":  year == currentYear,
		"AvailableYears": availableYears,
		"OverallAvg":     metadata["OverallAvg"],
		"SenateAvg":      metadata["SenateAvg"],
		"HouseAvg":       metadata["HouseAvg"],
		"TopLeader":      metadata["TopLeader"],
		"WorstLeader":    metadata["WorstLeader"],
		"TopAttendees":   top5,
		"WorstStreaks":   worst5,
		"RecentMissers":  recentlyMissing,
		"AllScores":      allScoresTable,
		"HasFilters":     hasFilters,
		"TotalCount":     metadata["TotalCount"],
		"FilteredCount":  len(filtered),
		"ActiveChamber":  chamber,
		"ActiveTenure":   tenure,
		"ActiveAge":      age,
		"Chambers":       chambers,
		"Tenures":        tenures,
		"Ages":           ages,
		"VoteFrequency":  voteFreq,
	}

	return c.Render("attendance_year", bindMap, "layouts/main")
}

func GetHtmxAttendanceScoreboard(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	yearStr := c.Params("year")
	var year int
	fmt.Sscanf(yearStr, "%d", &year)
	currentYear := time.Now().Year()
	if year < 2021 {
		year = currentYear
	}

	qs := string(c.Request().URI().QueryString())
	if qs != "" {
		c.Set("HX-Replace-Url", fmt.Sprintf("/attendance/%d?%s", year, qs))
	} else {
		c.Set("HX-Replace-Url", fmt.Sprintf("/attendance/%d", year))
	}

	chamber := c.Query("chamber", "All")
	tenure := c.Query("tenure", "All")
	age := c.Query("age", "All")

	filtered, metadata, err := computeAttendanceScores(db, year, chamber, tenure, age)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Database error: %v", err))
	}

	topAttendees := make([]ScoreboardItem, len(filtered))
	copy(topAttendees, filtered)
	sort.Slice(topAttendees, func(i, j int) bool {
		if topAttendees[i].AttendanceRate == topAttendees[j].AttendanceRate {
			return topAttendees[i].TotalVotes > topAttendees[j].TotalVotes
		}
		return topAttendees[i].AttendanceRate > topAttendees[j].AttendanceRate
	})
	limitTop := 5
	if len(topAttendees) < limitTop {
		limitTop = len(topAttendees)
	}
	top5 := topAttendees[:limitTop]

	worstStreaks := make([]ScoreboardItem, len(filtered))
	copy(worstStreaks, filtered)
	sort.Slice(worstStreaks, func(i, j int) bool {
		if worstStreaks[i].MaxMissedStreak == worstStreaks[j].MaxMissedStreak {
			return worstStreaks[i].AttendanceRate < worstStreaks[j].AttendanceRate
		}
		return worstStreaks[i].MaxMissedStreak > worstStreaks[j].MaxMissedStreak
	})
	limitWorst := 5
	if len(worstStreaks) < limitWorst {
		limitWorst = len(worstStreaks)
	}
	worst5 := worstStreaks[:limitWorst]

	limitRecent := 20
	var recentlyMissing []ScoreboardItem
	if year == currentYear {
		for _, s := range filtered {
			if s.RecentMissedVotes > 0 {
				recentlyMissing = append(recentlyMissing, s)
			}
		}
		sort.Slice(recentlyMissing, func(i, j int) bool {
			if recentlyMissing[i].RecentMissedVotes == recentlyMissing[j].RecentMissedVotes {
				return recentlyMissing[i].AttendanceRate < recentlyMissing[j].AttendanceRate
			}
			return recentlyMissing[i].RecentMissedVotes > recentlyMissing[j].RecentMissedVotes
		})
		if len(recentlyMissing) > limitRecent {
			recentlyMissing = recentlyMissing[:limitRecent]
		}
	}

	allScoresTable := make([]ScoreboardItem, len(filtered))
	copy(allScoresTable, filtered)
	sort.Slice(allScoresTable, func(i, j int) bool {
		if allScoresTable[i].AttendanceRate == allScoresTable[j].AttendanceRate {
			return allScoresTable[i].TotalVotes < allScoresTable[j].TotalVotes
		}
		return allScoresTable[i].AttendanceRate < allScoresTable[j].AttendanceRate
	})

	hasFilters := chamber != "All" || tenure != "All" || age != "All"

	return c.Render("attendance_rows", fiber.Map{
		"SelectedYear":  year,
		"IsCurrentYear": year == currentYear,
		"TopAttendees":  top5,
		"WorstStreaks":  worst5,
		"RecentMissers": recentlyMissing,
		"AllScores":     allScoresTable,
		"HasFilters":    hasFilters,
		"TotalCount":    metadata["TotalCount"],
		"FilteredCount": len(filtered),
	})
}
