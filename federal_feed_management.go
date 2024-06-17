package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

var rssLinks = []string{
	"https://www.govinfo.gov/rss/bills.xml",
}

type LawRssItem struct {
	FullTextUrl        string
	DescriptiveMetaUrl string
	Title              string
	Link               string
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
		db.Create(&newItem)
		return true, newItem
	}

	return false, newItem
}

func main() {

}

func DoBigApp() {
	db, err := setupDB()
	if err != nil {
		fmt.Println("Failed to setup database:", err)
		return
	}

	ch := make(LawRssItemChannel)
	go handleLawRss(rssLinks[0], ch)

	for item := range ch {
		fmt.Println(item)

		created, item := CreateDatabaseItemFromRssItem(item, db)
		if !created {
			fmt.Println("Item already exists in database")
			continue
		}

		text := downloadLawFullText(item.FullTextUrl)

		db.Create(&GovtLawText{
			GovtRssItemId: item.ID,
			Text:          text,
		})

		for _, chunk := range ChunkTextIntoTokenBlocks(text, 1500, 500) {
			response, err := CallGroqChatApi(Mixtral_8x7b, GetPrompt().PromptText, chunk)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			var tagData struct {
				Topics []string `json:"topics"`
			}

			err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &tagData)
			if err != nil {
				fmt.Println("Failed to unmarshal tag data:", err)
				return
			}

			for _, tagName := range tagData.Topics {
				tag := Tag{Name: tagName}
				db.FirstOrCreate(&tag, tag)

				tagRel := GovtRssItemTag{
					GovtRssItemId: item.ID,
					TagId:         tag.ID,
				}

				db.Create(&tagRel)
				fmt.Println("ADDED TAG", tagName, "---> ", tagRel.ID, tagRel.GovtRssItemId, tagRel.TagId)
			}

			fmt.Println("Tokens Consumed", response.Usage.TotalTokens, response.Usage.PromptTokens, response.Usage.CompletionTokens)
		}
	}

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

func handleLawRss(rssLink string, ch LawRssItemChannel) {
	defer close(ch)
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
		DescriptiveUrl := findHTMLTagWithText(item.Description, "Descriptive Metadata (MODS)")

		// Download the file associated with the HTML a tag
		// err = downloadFile(htmlTag.Href)
		// if err != nil {
		// 	fmt.Println("Failed to download file:", err)
		// 	return
		// }
		// fmt.Println(FullTextUrl, DescriptiveUrl)

		datetime, err := ParseDateTimeRssRobustly(item.Published)
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
