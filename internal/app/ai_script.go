package app

import (
	"github.com/hschmale16/the_final_stockbot/internal/m"
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
		panic(err)
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

	x, err := CallGroqChatApi(Gemma_7b, SYS, name)
	if err != nil {
		panic(err)
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
