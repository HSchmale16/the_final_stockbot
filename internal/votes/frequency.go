package votes

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DayVoteFrequency represents voting activity for a single calendar day
type DayVoteFrequency struct {
	DateStr     string  `json:"date_str"`     // e.g. "2024-01-15"
	DisplayDate string  `json:"display_date"` // e.g. "Jan 15, 2024"
	DayNumber   int     `json:"day_number"`   // 1..31
	Count       int     `json:"count"`        // count of votes on this day
	HeightPct   int     `json:"height_pct"`   // 0..100 percentage for sparkline bar height
	HasVotes    bool    `json:"has_votes"`    // true if Count > 0
	Tooltip     string  `json:"tooltip"`      // e.g. "Jan 15: 6 votes" or "Jan 15: No votes"
	SvgX        float64 `json:"svg_x"`
	SvgY        float64 `json:"svg_y"`
	SvgWidth    float64 `json:"svg_width"`
	SvgHeight   float64 `json:"svg_height"`
}

// MonthVoteFrequency represents voting activity across all days in a calendar month
type MonthVoteFrequency struct {
	MonthNumber int                `json:"month_number"` // 1..12
	MonthName   string             `json:"month_name"`   // "Jan", "Feb", ...
	TotalVotes  int                `json:"total_votes"`  // total votes in this month
	DividerX    float64            `json:"divider_x"`
	Days        []DayVoteFrequency `json:"days"` // all days in this month
}

// YearVoteFrequency encapsulates the yearly vote frequency breakdown by month and day
type YearVoteFrequency struct {
	Year         int                  `json:"year"`
	TotalVotes   int                  `json:"total_votes"`
	MaxDaily     int                  `json:"max_daily"`
	BusiestDate  string               `json:"busiest_date"`
	BusiestCount int                  `json:"busiest_count"`
	VotingDays   int                  `json:"voting_days"`
	Months       []MonthVoteFrequency `json:"months"`
}

type dailyVoteCountRow struct {
	VoteDate string `gorm:"column:vote_date"`
	Count    int    `gorm:"column:count"`
}

// computeYearVoteFrequency calculates the daily vote counts and sparkline coordinates for all 12 months in the given year using SQL GROUP BY.
func computeYearVoteFrequency(db *gorm.DB, year int, chamber string) YearVoteFrequency {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	var rows []dailyVoteCountRow

	query := db.Table("votes").
		Select("TO_CHAR(action_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS vote_date, COUNT(*) AS count").
		Where("action_at BETWEEN ? AND ?", start, end)

	if chamber != "" && chamber != "All" {
		query = query.Where("chamber ILIKE ?", chamber)
	}

	_ = query.
		Group("TO_CHAR(action_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Order("vote_date ASC").
		Scan(&rows)

	dailyCounts := make(map[string]int, len(rows))
	maxDaily := 0
	busiestDate := ""
	busiestCount := 0
	totalVotes := 0
	votingDays := len(rows)

	for _, r := range rows {
		dailyCounts[r.VoteDate] = r.Count
		totalVotes += r.Count
		if r.Count > maxDaily {
			maxDaily = r.Count
			busiestDate = r.VoteDate
			busiestCount = r.Count
		}
	}

	// Calculate total days in the year (365 or 366)
	totalYearDays := 0
	for m := 1; m <= 12; m++ {
		totalYearDays += time.Date(year, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Day()
	}

	const svgWidth = 1000.0
	const svgHeight = 60.0
	const baselineY = 56.0
	const maxBarH = 50.0

	slotWidth := svgWidth / float64(totalYearDays)
	barWidth := slotWidth * 0.72

	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	months := make([]MonthVoteFrequency, 0, 12)
	globalDayIdx := 0

	for m := 1; m <= 12; m++ {
		numDays := time.Date(year, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Day()
		days := make([]DayVoteFrequency, 0, numDays)
		monthTotal := 0

		for d := 1; d <= numDays; d++ {
			dt := time.Date(year, time.Month(m), d, 0, 0, 0, 0, time.UTC)
			dStr := dt.Format("2006-01-02")
			displayDate := dt.Format("Jan 02, 2006")
			count := dailyCounts[dStr]
			monthTotal += count

			heightPct := 0
			var svgH, svgY float64
			if count > 0 {
				if maxDaily > 0 {
					heightPct = int(float64(count) / float64(maxDaily) * 100.0)
					if heightPct < 15 {
						heightPct = 15 // minimum visible height for sparkline bar
					}
					svgH = float64(count) / float64(maxDaily) * maxBarH
					if svgH < 5.0 {
						svgH = 5.0
					}
				} else {
					heightPct = 15
					svgH = 5.0
				}
				svgY = baselineY - svgH
			} else {
				svgH = 2.0
				svgY = baselineY - 2.0
			}

			svgX := float64(globalDayIdx)*slotWidth + (slotWidth-barWidth)/2.0

			voteWord := "votes"
			if count == 1 {
				voteWord = "vote"
			}
			tooltip := fmt.Sprintf("%s: %d %s", displayDate, count, voteWord)
			if count == 0 {
				tooltip = fmt.Sprintf("%s: No votes", displayDate)
			}

			days = append(days, DayVoteFrequency{
				DateStr:     dStr,
				DisplayDate: displayDate,
				DayNumber:   d,
				Count:       count,
				HeightPct:   heightPct,
				HasVotes:    count > 0,
				Tooltip:     tooltip,
				SvgX:        svgX,
				SvgY:        svgY,
				SvgWidth:    barWidth,
				SvgHeight:   svgH,
			})
			globalDayIdx++
		}

		dividerX := 0.0
		if m < 12 {
			dividerX = float64(globalDayIdx) * slotWidth
		}

		months = append(months, MonthVoteFrequency{
			MonthNumber: m,
			MonthName:   monthNames[m],
			TotalVotes:  monthTotal,
			DividerX:    dividerX,
			Days:        days,
		})
	}

	busiestDisplay := ""
	if busiestDate != "" {
		if t, err := time.Parse("2006-01-02", busiestDate); err == nil {
			busiestDisplay = t.Format("Jan 02")
		}
	}

	return YearVoteFrequency{
		Year:         year,
		TotalVotes:   totalVotes,
		MaxDaily:     maxDaily,
		BusiestDate:  busiestDisplay,
		BusiestCount: busiestCount,
		VotingDays:   votingDays,
		Months:       months,
	}
}
