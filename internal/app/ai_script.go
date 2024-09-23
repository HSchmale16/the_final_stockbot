package app

import (
	"log"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	henry_groq "github.com/hschmale16/the_final_stockbot/pkg/groq"
	"golang.org/x/exp/slices"
)

var TagCssColorTypes = []string{
	"bg-person",
	"bg-law",
	"bg-agency",
	"bg-organization",
	"bg-geography",
	"bg-country",
	"bg-unknown",
	"bg-secondary",
}

func DoTagUpdates() {
	db, err := m.SetupDB()
	if err != nil {
		log.Fatal("Failed to setup db" + err.Error())
	}

	var tags []m.Tag
	db.Limit(18).Where("css_color = ?", "bg-secondary").Or("short_line = ?", "").Order("RANDOM()").Find(&tags)

	for _, tag := range tags {
		if tag.CssColor == "bg-secondary" {
			cat := GetAiDescription(tag.Name)
			tag.CssColor = cat
		}

		if tag.ShortLine == "" {
			tag.ShortLine = GetAiOneLine(tag.Name)
		}

		db.Save(&tag)
	}
}

func GetAiOneLine(name string) string {
	SYS := `The user will provide a tag. You must describe what it means and provide a very short description about what it is. Keep it extremely short.`

	x, err := henry_groq.CallGroqChatApi(henry_groq.Gemma_7b, SYS, name)
	if err != nil {
		log.Fatal("Failed to describe tag: " + name + " " + err.Error())
	}

	return x.Choices[0].Message.Content
}

func GetAiDescription(name string) string {
	SYS := `The user will provide a tag. You must categorize it according to the following categories: 
	"bg-person",
	"bg-law",
	"bg-agency",
	"bg-organization",
	"bg-geography",
	"bg-country",
	"bg-unknown",
	"bg-secondary"

Print only the category.

For example, israel is a country
The president is a person.`

	x, err := henry_groq.CallGroqChatApi(henry_groq.Gemma_7b, SYS, name)
	if err != nil {
		log.Fatal("Failed to categorize tag: " + name + " " + err.Error())
	}

	content := x.Choices[0].Message.Content

	if slices.Contains(TagCssColorTypes, content) {
		return content
	}

	return "bg-unknown"
}
