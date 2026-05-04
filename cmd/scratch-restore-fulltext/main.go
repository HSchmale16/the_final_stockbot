package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hschmale16/the_final_stockbot/internal/congress"
	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func downloadLawFullText(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error downloading full text:", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading full text:", err)
		return ""
	}
	return string(body)
}

func main() {
	db, err := m.SetupDB()
	if err != nil {
		panic(err)
	}

	// 1. Clear linkages
	if err := db.Exec("DELETE FROM hearing_attended_members").Error; err != nil {
		log.Fatal("Failed to clear hearing_attended_members:", err)
	}
	fmt.Println("Cleared hearing_attended_members linkages")

	// 2. Restore FullText
	var hearings []congress.Hearing
	// Find hearings where we likely cleared full text (it has a URL but no text)
	if err := db.Where("full_text = '' AND full_text_url != ''").Find(&hearings).Error; err != nil {
		log.Fatal("Failed to query hearings:", err)
	}

	fmt.Printf("Found %d hearings to restore FullText for\n", len(hearings))

	for i, h := range hearings {
		text := downloadLawFullText(h.FullTextUrl)
		if text != "" {
			h.FullText = text
			if err := db.Save(&h).Error; err != nil {
				log.Println("Error saving hearing:", h.ID, err)
			}
			if i%50 == 0 {
				fmt.Printf("Restored %d / %d\n", i+1, len(hearings))
			}
		}
	}
	fmt.Println("Done restoring FullText.")
}
