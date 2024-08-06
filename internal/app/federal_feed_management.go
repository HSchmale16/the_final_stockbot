package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"

	. "github.com/hschmale16/the_final_stockbot/internal/m"
	henry_groq "github.com/hschmale16/the_final_stockbot/pkg/groq"
)

var RssLinks = []string{
	"https://www.govinfo.gov/rss/bills.xml",
	"https://www.govinfo.gov/rss/bills-enr.xml",
	"https://www.govinfo.gov/rss/plaw.xml",
}

var FederalRegisterFeed = "https://www.govinfo.gov/rss/fr.xml"

type LawRssItem struct {
	FullTextUrl        string
	DescriptiveMetaUrl string
	Title              string
	Link               string
	Category           []string
	PubDate            time.Time
}

type LawRssItemChannel chan LawRssItem

func CreateDatabaseItemFromRssItem(item LawRssItem, db *gorm.DB) (bool, GovtRssItem) {
	newItem := GovtRssItem{
		DescriptiveMetaUrl: item.DescriptiveMetaUrl,
		FullTextUrl:        item.FullTextUrl,
		Title:              item.Title,
		Link:               item.Link,
		PubDate:            item.PubDate,
	}

	// Search by link
	var count int64
	db.Model(&GovtRssItem{}).Where("link = ?", item.Link).Count(&count)

	log.Println("Count:", count, "Link:", item.Link)
	if count == 0 {
		// for _, tagName := range item.Category {
		// 	tag := GetTag(db, tagName)
		// 	newItem.Categories = append(newItem.Categories, tag)
		// }

		fmt.Println(newItem)
		x := db.Debug().Create(&newItem)
		if x.Error != nil {
			log.Fatal("Failed to create item:", x.Error)
		}

		return true, newItem
	}

	return false, newItem
}

func RunFetcherService(ch LawRssItemChannel) {
	defer close(ch)
	db, err := SetupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	for item := range ch {
		fmt.Println(item)

		created, item := CreateDatabaseItemFromRssItem(item, db)
		if !created {
			fmt.Println("Item already exists in database")
			continue
		}

		text := downloadLawFullText(item.FullTextUrl)
		mods := downloadModsXML(item.DescriptiveMetaUrl)

		db.Create(&GovtLawText{
			GovtRssItemId: item.ID,
			Text:          text,
			ModsXML:       mods,
		})

		modsData := ReadLawModsData(mods)
		congressMembers := make([]string, len(modsData.CongressMembers))
		for i, member := range modsData.CongressMembers {
			congressMembers[i] = member.Name
		}
		CreateTagsOnItem(congressMembers, item, 0, db)
		ScanLawSponsors(modsData, item, db)
		ProcessLawTextForTags(item, db)
	}
}

func ScanLawSponsors(modsData LawModsData, item GovtRssItem, db *gorm.DB) {
	item.Metadata = modsData
	db.Save(&item)

	// Find the congress member
	for _, congMember := range modsData.CongressMembers {
		var dbCongMember DB_CongressMember
		db.Where("bio_guide_id = ?", congMember.BioGuideId).First(&dbCongMember)
		if dbCongMember.BioGuideId == "" {
			log.Printf("Could not find congress member %s\n", congMember)
			continue
		}
		// Create the association
		dbCongMemberSponsored := CongressMemberSponsored{
			DB_CongressMemberBioGuideId: congMember.BioGuideId,
			GovtRssItemId:               item.ID,
			// Slightly denormalized here. But it makes sense for the kind of questions we are asking and
			// it can change over time. Trust the library of congress to get it right
			Chamber:        congMember.Chamber,
			CongressNumber: congMember.Congress,
			Role:           congMember.Role,
		}

		x := CongressMemberSponsored{
			DB_CongressMemberBioGuideId: congMember.BioGuideId,
			GovtRssItemId:               item.ID,
		}

		count := int64(0)
		db.Where(&x).Count(&count)

		if count == 0 {
			db.Create(&dbCongMemberSponsored)
		}

		//fmt.Println("SCANNED SPONSORS ADDED: ", result.RowsAffected)
	}

	for _, committee := range modsData.CongressCommittees {
		fmt.Println(committee)

		committee.AuthorityId = strings.TrimSuffix(committee.AuthorityId, "00")

		// Find the committee
		var dbCommittee DB_CongressCommittee
		err := db.Debug().Where("LOWER(thomas_id) = ?", committee.AuthorityId).First(&dbCommittee)
		if err.Error != nil {
			log.Printf("Could not find committee %s\n", committee.AuthorityId)
			continue
		}

		fmt.Println(item)

		//db.Debug().Model(&dbCommittee).Association("GovtRssItems").Append(&item)
	}
}

func FindUntaggedLaws() {
	db, err := SetupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	var items []GovtRssItem
	db.Debug().Where("id NOT IN (SELECT govt_rss_item_id FROM govt_rss_item_tag)").Find(&items)

	fmt.Println("Found", len(items), "untagged items")

	ch := make(chan GovtRssItem, 100)
	defer close(ch)
	go ProcessLawItemsFromChannel(ch)

	for _, item := range items {
		ch <- item
	}
}

func ProcessLawItemsFromChannel(ch chan GovtRssItem) {
	db, err := SetupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	for item := range ch {
		fmt.Println("RE-Processing item:", item.ID)
		ProcessLawTextForTags(item, db)
	}
}

func ProcessLawTextForTags(src GovtRssItem, db *gorm.DB) {
	var item GovtLawText
	db.First(&item, "govt_rss_item_id = ?", src.ID)

	var count int64
	db.Model(&GovtRssItemTag{}).Where("govt_rss_item_id = ?", src.ID).Count(&count)

	fmt.Println("Target item already has tags:", count)

	if src.Metadata.IsAppropriation {
		CreateTagsOnItem([]string{"Appropriation"}, src, 0, db)
	}

	textOffset := 0
	for _, chunk := range ChunkTextIntoTokenBlocks(item.Text, 1000, 500) {
		var response henry_groq.GroqChatCompletion
		var err error
		for i := 0; i < 3; i++ {
			model := henry_groq.Llama3_8B
			response, err = henry_groq.CallGroqChatApi(model, GetPrompt().PromptText, chunk)
			if err == nil {
				break
			}
			fmt.Println("Error:", err)
			db.Create(&GenerationError{
				Model:         string(model),
				ErrorMessage:  err.Error(),
				AttemptedText: chunk,
			})

			time.Sleep(15 * time.Second)
		}

		var tagData struct {
			Topics []string `json:"topics"`
		}

		body := []byte(response.Choices[0].Message.Content)
		err = json.Unmarshal(body, &tagData)
		if err != nil {
			// Try to reparse if the response ends in a single ] character
			if string(body[len(body)-1]) == "]" {
				body = append(body, byte('}'))
				err = json.Unmarshal(body, &tagData)
				if err != nil {
					fmt.Println("Failed to unmarshal fixed repsonse", err)
				}
			}
		}

		CreateTagsOnItem(tagData.Topics, src, textOffset, db)

		fmt.Println("Tokens Consumed", response.Usage.TotalTokens, response.Usage.PromptTokens, response.Usage.CompletionTokens)
		textOffset += len(chunk)
	}
}

func CreateTagsOnItem(tags []string, item GovtRssItem, textOffset int, db *gorm.DB) {
	for _, tag := range tags {
		tag := GetTag(db, tag)

		// Check if the tag relationship already exists
		var count int64
		db.Model(&GovtRssItemTag{}).Where("govt_rss_item_id = ? AND tag_id = ?", item.ID, tag.ID).Count(&count)

		var myGovt = GovtRssItemTag{
			GovtRssItemId: item.ID,
			TagId:         tag.ID,
		}

		db.FirstOrCreate(&myGovt, myGovt)
		db.Create(&LawOffset{
			GovtRssItemTagId: myGovt.ID,
			Offset:           textOffset,
		})
	}
}

func downloadModsXML(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to make GET request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Read the body and return the string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read body:", err)
		return ""
	}

	return string(body)
}

func downloadLawFullText(url string) string {
	// Make a get request and return the text between the first pretag.

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to make GET request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Parse the HTML data
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Failed to parse HTML data:", err)
		return ""
	}

	// Find the first pre tag and return the text
	return doc.Find("pre").First().Text()
}

func HandleLawRss(rssLink string, ch LawRssItemChannel) {
	// Parse the RSS feed using gofeed
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(rssLink)
	if err != nil {
		fmt.Println("Failed to parse RSS feed:", err)
		return
	}

	// Print the title and description of each item in the feed
	for _, item := range feed.Items {
		// fmt.Println("Title:", item.Title)
		// fmt.Println("Date: ", item.Published)
		// fmt.Println("Link: ", item.Link)
		// fmt.Println("Description:", item.Description)

		// Find the HTML a tag with the text "TEXT"
		FullTextUrl := findHTMLTagWithText(item.Description, "TEXT")
		// If the string is empty find the XML URL For Federal Register Stuff
		if len(FullTextUrl) == 0 {
			FullTextUrl = findHTMLTagWithText(item.Description, "XML")
		}

		DescriptiveUrl := findHTMLTagWithText(item.Description, "Descriptive Metadata (MODS)")

		// Download the file associated with the HTML a tag
		// err = downloadFile(htmlTag.Href)
		// if err != nil {
		// 	fmt.Println("Failed to download file:", err)
		// 	return
		// }
		// fmt.Println(FullTextUrl, DescriptiveUrl)

		datetime, err := henry_groq.ParseDateTimeRssRobustly(item.Published)
		if err != nil {
			fmt.Println("Failed to parse datetime:", err)
			return
		}

		ch <- LawRssItem{
			FullTextUrl:        FullTextUrl,
			DescriptiveMetaUrl: DescriptiveUrl,
			Title:              item.Title,
			Link:               item.Link,
			PubDate:            datetime,
			Category:           item.Categories,
		}
	}

}

/**
 * Find the HTML a tag with the specified text
 * @param htmlData The partial HTML description to search
 * @param linkText the link text to search for.
 * @return The href value of the targetted link tag or an empty string if not found.
 */
func findHTMLTagWithText(htmlData, linkText string) string {
	// Parse the HTML data
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlData))
	if err != nil {
		fmt.Println("Failed to parse HTML data:", err)
		return ""
	}

	// Find the a tag with the specified text and return the href value
	// use linear search
	for _, link := range doc.Find("a").Nodes {
		for _, attr := range link.Attr {
			if attr.Key == "href" && strings.Contains(link.FirstChild.Data, linkText) {
				return attr.Val
			}
		}
	}

	return ""
}
