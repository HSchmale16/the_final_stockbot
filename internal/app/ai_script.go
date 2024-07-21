package app

import (
	"fmt"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"golang.org/x/exp/slices"
)

func DoTagUpdates() {
	db, err := m.SetupDB()
	if err != nil {
		panic(err)
	}

	var tags []m.Tag
	db.Limit(20).Order("RANDOM()").Find(&tags)

	for _, tag := range tags {
		fmt.Println(tag)
		if tag.CssColor == "bg-secondary" {
			cat := GetAiDescription(tag.Name)
			tag.CssColor = cat
		}

		db.Save(&tag)
	}
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

	x, err := CallGroqChatApi(Gemma_7b, SYS, name)
	if err != nil {
		panic(err)
	}

	content := x.Choices[0].Message.Content

	if slices.Contains(TagCssColorTypes, content) {
		return content
	}

	return "bg-unknown"
}
