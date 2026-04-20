package app

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
)

const (
	PREAMBLE = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url><loc>https://www.dirtycongress.com/</loc></url>
<url><loc>https://www.dirtycongress.com/help/faq</loc></url>
<url><loc>https://www.dirtycongress.com/tos</loc></url>
<url><loc>https://www.dirtycongress.com/travel</loc></url>
<url><loc>https://www.dirtycongress.com/congress-members</loc></url>
<url><loc>https://www.dirtycongress.com/committees</loc></url>
<url><loc>https://www.dirtycongress.com/hearings</loc></url>
`

	POSTAMBLE    = `</urlset>`
	URL_TEMPLATE = "<url><loc>%s</loc><lastmod>%s</lastmod></url>\n"

	SITEMAP_DT_FORMAT = "2006-01-02T15:04:05-07:00"
)

// File Location
var fileLocFlag string

func init() {
	flag.StringVar(&fileLocFlag, "fileLoc", "/var/lib/final_stockbot/the_final_stockbot/static/sitemap.xml", "Location of the sitemap.xml file")
}

func MakeSitemap() {
	db, err := m.SetupDB()
	if err != nil {
		panic(err)
	}

	// Open ~/the_final_stockbot/static/sitemap.xml for writing text
	// Wordexpansion must be perfomred
	file, err := os.Create(fileLocFlag)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write the preamble
	file.WriteString(PREAMBLE)

	SITEURL := "https://www.dirtycongress.com"

	// Laws are hidden now as per user request
	/*
		rows, err := db.Model(&GovtRssItem{}).Select("id, title, link, pub_date").Order("pub_date DESC").Rows()
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var govtRssItem GovtRssItem
			db.ScanRows(rows, &govtRssItem)

			// Write the url
			url := SITEURL + "/law/" + strconv.Itoa(int(govtRssItem.ID))
			tmp := fmt.Sprintf(URL_TEMPLATE, url, govtRssItem.PubDate.Format(SITEMAP_DT_FORMAT))
			file.WriteString(tmp)
		}
	*/

	rows, err := db.Model(&DB_CongressMember{}).Select("bio_guide_id, congress_member_info").Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	today := time.Now()
	for rows.Next() {
		var congressMember DB_CongressMember
		db.ScanRows(rows, &congressMember)

		// Write the url
		if congressMember.IsActiveMember() {
			url := SITEURL + "/congress-member/" + congressMember.BioGuideId
			tmp := fmt.Sprintf(URL_TEMPLATE, url, today.Format(SITEMAP_DT_FORMAT))
			file.WriteString(tmp)
		}
	}

	// Add Committees
	rows, err = db.Model(&DB_CongressCommittee{}).Select("thomas_id").Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var committee DB_CongressCommittee
		db.ScanRows(rows, &committee)
		url := SITEURL + "/committee/" + committee.ThomasId
		tmp := fmt.Sprintf(URL_TEMPLATE, url, today.Format(SITEMAP_DT_FORMAT))
		file.WriteString(tmp)
	}

	file.WriteString(POSTAMBLE)
}
