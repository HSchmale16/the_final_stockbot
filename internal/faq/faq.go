package faq

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

//go:embed faq.yml
var faqYaml string

type FaqCategory struct {
	Category  string `yaml:"category"`
	Questions []struct {
		Question string `yaml:"question"`
		Answer   string `yaml:"answer"`
		Slug     string `yaml:"slug"`
	}
}

var faqCategories []FaqCategory

func init() {
	yaml.Unmarshal([]byte(faqYaml), &faqCategories)
}

func faq(c *fiber.Ctx) error {
	return c.Render("faq", fiber.Map{
		"Categories":  faqCategories,
		"Description": "Frequently Asked Questions about DirtyCongress.com",
		"Title":       "Frequently Asked Questions - DirtyCongress.com",
	}, "layouts/main")
}

func SetupRoutes(app *fiber.App) {
	app.Get("/help/faq", faq)

	app.Get("/help/about-congress", func(c *fiber.Ctx) error {
		return c.Render("help", fiber.Map{}, "layouts/main")
	})
}
