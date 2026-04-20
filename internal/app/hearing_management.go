package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	henry_groq "github.com/hschmale16/the_final_stockbot/pkg/groq"
)

var HearingRssFeed = "https://www.govinfo.gov/rss/chrg.xml"

func RunHearingFetcherService() {
	db, err := m.SetupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(HearingRssFeed)
	if err != nil {
		fmt.Println("Failed to parse RSS feed:", err)
		return
	}

	for _, item := range feed.Items {
		// Find URLs
		FullTextUrl := findHTMLTagWithText(item.Description, "TEXT")
		if len(FullTextUrl) == 0 {
			FullTextUrl = findHTMLTagWithText(item.Description, "XML")
		}
		ModsUrl := findHTMLTagWithText(item.Description, "Descriptive Metadata (MODS)")

		if ModsUrl == "" {
			log.Println("No MODS URL found for hearing:", item.Title)
			continue
		}

		datetime, err := henry_groq.ParseDateTimeRssRobustly(item.Published)
		if err != nil {
			fmt.Println("Failed to parse datetime:", err)
			continue
		}

		// Use the core processing function
		processHearing(db, item.Title, item.Link, datetime, ModsUrl, FullTextUrl)
	}
}

func RunHearingBackfill(congressNum int) {
	db, err := m.SetupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	// 119th is 2025-2026
	startYear := 1789 + (congressNum-1)*2
	years := []int{startYear, startYear + 1}

	// Also include current year if it's not in the range
	currentYear := time.Now().Year()
	found := false
	for _, y := range years {
		if y == currentYear {
			found = true
		}
	}
	if !found {
		years = append(years, currentYear)
	}

	for _, year := range years {
		sitemapUrl := fmt.Sprintf("https://www.govinfo.gov/sitemap/CHRG_%d_sitemap.xml", year)
		fmt.Println("Processing sitemap:", sitemapUrl)

		resp, err := http.Get(sitemapUrl)
		if err != nil {
			fmt.Println("Failed to fetch sitemap:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Sitemap not found for year:", year)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read sitemap body:", err)
			continue
		}

		// Find all CHRG-XXX items for this congress
		prefix := fmt.Sprintf("CHRG-%d", congressNum)
		// Match loc tags: <loc>https://www.govinfo.gov/app/details/CHRG-119hhrg62598</loc>
		re := regexp.MustCompile(`<loc>https://www.govinfo.gov/app/details/(` + prefix + `[^<]+)</loc>`)
		matches := re.FindAllStringSubmatch(string(body), -1)

		fmt.Printf("Found %d potential hearings for %d in year %d\n", len(matches), congressNum, year)

		for _, match := range matches {
			pkgId := match[1]
			link := "https://www.govinfo.gov/app/details/" + pkgId
			modsUrl := fmt.Sprintf("https://www.govinfo.gov/metadata/pkg/%s/mods.xml", pkgId)

			// Check if already processed
			var existing Hearing
			db.Where("link = ?", link).First(&existing)
			if existing.ID != 0 && existing.FullText != "" && existing.PdfUrl != "" {
				continue
			}

			fmt.Println("Backfilling Hearing:", pkgId)
			processHearing(db, "", link, time.Now(), modsUrl, "")
		}
	}
}

func processHearing(db *gorm.DB, title string, link string, pubDate time.Time, modsUrl string, fullTextUrl string) {
	// Check if exists
	var existing Hearing
	db.Where("link = ?", link).First(&existing)
	if existing.ID != 0 && existing.FullText != "" && existing.PdfUrl != "" {
		return
	}

	// Download MODS
	modsXml := downloadModsXML(modsUrl)
	modsData := m.ReadHearingModsData(modsXml)

	// Fallback for title if empty
	if title == "" {
		title = modsData.Title
	}

	// Fallback for full text url if empty
	if fullTextUrl == "" {
		fullTextUrl = modsData.FullTextUrl
	}

	var text string
	if fullTextUrl != "" {
		text = downloadLawFullText(fullTextUrl)
	}

	witnessesJson, _ := json.Marshal(modsData.Witnesses)

	hearing := Hearing{
		Title:       title,
		Link:        link,
		PubDate:     pubDate,
		HeldDate:    modsData.HeldDate,
		FullTextUrl: fullTextUrl,
		PdfUrl:      modsData.PdfUrl,
		ModsUrl:     modsUrl,
		FullText:    text,
		Witnesses:   witnessesJson,
	}

	if existing.ID != 0 {
		hearing.ID = existing.ID
		db.Save(&hearing)
	} else {
		err := db.Create(&hearing).Error
		if err != nil {
			log.Println("Failed to create hearing:", err)
			return
		}

		// Add Members
		for _, member := range modsData.CongressMembers {
			var dbMember m.DB_CongressMember
			db.Where("bio_guide_id = ?", member.BioGuideId).First(&dbMember)
			if dbMember.BioGuideId != "" {
				db.Model(&hearing).Association("Members").Append(&dbMember)
			}
		}

		// Add Committees
		for _, committee := range modsData.CongressCommittees {
			authorityId := strings.TrimSuffix(committee.AuthorityId, "00")
			var dbCommittee m.DB_CongressCommittee
			db.Where("LOWER(thomas_id) = ?", strings.ToLower(authorityId)).First(&dbCommittee)
			if dbCommittee.ThomasId != "" {
				db.Model(&hearing).Association("Committees").Append(&dbCommittee)
			}
		}
	}
}
