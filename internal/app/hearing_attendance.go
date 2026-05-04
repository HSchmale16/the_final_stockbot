package app

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hschmale16/the_final_stockbot/internal/congress"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

var rePresent = regexp.MustCompile(`(?i)\b(?:Members present|Present|Committee Members present)\s*:(.*?)(?:\.|$)`)

var titleRemovals = []string{
	"Senators", "Senator",
	"Representatives", "Representative",
	"Hon.", "Hon",
	"Mr.", "Mr",
	"Mrs.", "Mrs",
	"Ms.", "Ms",
	"Chairman", "Chairwoman", "Chair",
	"Ranking Member",
}

// ExtractAttendees extracts a list of raw last names from a transcript string
func ExtractAttendees(fullText string) []string {
	// clean up newlines for the match
	cleanedText := strings.ReplaceAll(fullText, "\n", " ")
	spaceRe := regexp.MustCompile(`\s+`)
	cleanedText = spaceRe.ReplaceAllString(cleanedText, " ")

	matches := rePresent.FindStringSubmatch(cleanedText)
	if len(matches) < 2 {
		return nil
	}

	rawNamesStr := matches[1]

	// Remove common titles
	for _, title := range titleRemovals {
		// Use regex for word boundary replacement to avoid replacing inside names (unlikely but safe)
		reTitle := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(title) + `\b`)
		rawNamesStr = reTitle.ReplaceAllString(rawNamesStr, "")
	}

	// Split by commas and "and"
	// First replace " and " with ","
	rawNamesStr = strings.ReplaceAll(rawNamesStr, " and ", ",")
	rawNamesStr = strings.ReplaceAll(rawNamesStr, " & ", ",")

	parts := strings.Split(rawNamesStr, ",")
	var finalNames []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" && len(name) > 1 { // avoid stray punctuation
			finalNames = append(finalNames, name)
		}
	}

	return finalNames
}

func parseYearFromHeldDate(heldDate string) int {
	if len(heldDate) >= 4 {
		yearStr := heldDate[:4]
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			return year
		}
	}
	return 0
}

func MatchMembers(activeMembers []m.DB_CongressMember, rawNames []string, hearing *congress.Hearing) []m.DB_CongressMember {
	if len(rawNames) == 0 {
		return nil
	}

	var matchedMembers []m.DB_CongressMember
	seenIds := make(map[string]bool)

	for _, rawName := range rawNames {
		searchName := strings.ToLower(rawName)

		// 1. Try to match against the hearing's existing 'Members' (Committee members) first
		var bestMatch *m.DB_CongressMember
		for i, cm := range hearing.Members {
			last := strings.ToLower(cm.CongressMemberInfo.Name.Last)
			if strings.Contains(last, searchName) || strings.Contains(searchName, last) {
				bestMatch = &hearing.Members[i]
				break
			}
		}

		// 2. If no match in committee, check all active members for that year
		if bestMatch == nil {
			for i, am := range activeMembers {
				last := strings.ToLower(am.CongressMemberInfo.Name.Last)
				if strings.Contains(last, searchName) || strings.Contains(searchName, last) {
					bestMatch = &activeMembers[i]
					break
				}
			}
		}

		if bestMatch != nil && !seenIds[bestMatch.BioGuideId] {
			matchedMembers = append(matchedMembers, *bestMatch)
			seenIds[bestMatch.BioGuideId] = true
		}
	}

	return matchedMembers
}

func ProcessHearingAttendance(db *gorm.DB) error {
	var hearings []congress.Hearing
	
	// Preload Members (the committee members) so we can prioritize matching against them
	if err := db.Preload("Members").Where("full_text != ''").Find(&hearings).Error; err != nil {
		return fmt.Errorf("failed to query hearings: %w", err)
	}

	var allMembers []m.DB_CongressMember
	if err := db.Find(&allMembers).Error; err != nil {
		return fmt.Errorf("failed to load congress members: %w", err)
	}

	log.Printf("Found %d hearings to parse attendance for\n", len(hearings))

	processedCount := 0
	clearedCount := 0

	for _, h := range hearings {
		// We could skip if it already has AttendedMembers, but since it's an association table, 
		// it's easier to just check if we have any existing associations.
		var count int64
		db.Table("hearing_attended_members").Where("hearing_id = ?", h.ID).Count(&count)
		if count > 0 {
			// Already processed attendance
			continue
		}

		// Check if it's a PDF-only stub
		if strings.Contains(h.FullText, "REFER TO PDF") || strings.Contains(h.FullText, "TEXT NOT AVAILABLE") {
			h.IsPdfOnly = true
			db.Save(&h)
		}

		rawNames := ExtractAttendees(h.FullText)
		if len(rawNames) > 0 {
			year := parseYearFromHeldDate(h.HeldDate)
			if year == 0 {
				year = h.PubDate.Year()
			}
			
			var activeMembers []m.DB_CongressMember
			for _, member := range allMembers {
				if member.CongressMemberInfo.ServedDuringYear(year) {
					activeMembers = append(activeMembers, member)
				}
			}

			matchedMembers := MatchMembers(activeMembers, rawNames, &h)
			
			if len(matchedMembers) > 0 {
				err := db.Model(&h).Association("AttendedMembers").Append(matchedMembers)
				if err != nil {
					log.Printf("Failed to append attended members for hearing %d: %v\n", h.ID, err)
					continue
				}
				// Save to save the association
				
				processedCount++
			}
		}
		
		// Clear the full text to save space now that we've extracted what we need, 
		// or if we already flagged it as IsPdfOnly
		if h.FullText != "" {
			h.FullText = ""
			db.Save(&h)
			clearedCount++
		}
	}

	log.Printf("Finished parsing attendance. Successfully processed %d new hearings. Cleared full text from %d records.\n", processedCount, clearedCount)
	return nil
}
