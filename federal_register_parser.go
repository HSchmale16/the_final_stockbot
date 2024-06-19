package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

/**
* Helps wrangle federal register stuff for chunking and parsing into specific items and subsections
 */

func FederalRegisterParser() {
	// Open the FR-2024-06-20.xml file
	file, err := os.Open("/home/hschmale/src/the_final_stockbot/FR-2024-06-20.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file contents
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Parse into xml tree
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	_ = GetTag(db, "Federal Register")

	rules := doc.Find("RULES")
	rules.Find("RULE").Each(func(i int, s *goquery.Selection) {
		ProcessRule(s.Text(), "RULE", db)
	})

	doc.Find("PRORULES").Find("PRORULE").Each(func(i int, s *goquery.Selection) {
		ProcessRule(s.Text(), "PRORULE", db)
	})

	doc.Find("NOTICE").Each(func(i int, s *goquery.Selection) {
		ProcessRule(s.Text(), "NOTICE", db)
	})

	// fmt.Println(doc)
}

func ProcessRule(text, itemType string, db *gorm.DB) {
	// Parse the text into a rule
	// Create a new rule in the database
	// Return the rule
	item := FederalRegisterItem{
		FullText: text,
		Type:     itemType,
	}

	db.Create(&item)

	for _, chunk := range ChunkTextIntoTokenBlocks(text, 2500, 250) {
		var response GroqChatCompletion
		var err error
		for i := 0; i < 3; i++ {
			model := Llama3_8B
			response, err = CallGroqChatApi(model, GetPrompt().PromptText, chunk)
			if err == nil {
				break
			}
			fmt.Println("Error:", err)
			db.Create(&GenerationError{
				Model:         string(model),
				ErrorMessage:  err.Error(),
				AttemptedText: chunk,
				Source:        "ProcessFederalRegisterRule",
			})

			time.Sleep(10 * time.Second)
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

		for _, tagName := range tagData.Topics {
			tag := GetTag(db, tagName)
			item.Tags = append(item.Tags, tag)
		}
		db.Save(&item)

		fmt.Println("Tokens Consumed", response.Usage.TotalTokens, response.Usage.PromptTokens, response.Usage.CompletionTokens)
	}
}
