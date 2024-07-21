package app

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
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

func GetEditTagView(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tag m.Tag
	db.First(&tag, c.Params("tag_id"))

	return c.Render("htmx/tag_edit_form", fiber.Map{
		"Tag":           tag,
		"CssColorTypes": TagCssColorTypes,
	})
}

func PutTagUpdate(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tag m.Tag
	db.Debug().First(&tag, c.Params("tag_id"))

	newColor := c.FormValue("css_color")
	if slices.Contains(TagCssColorTypes, newColor) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid color type",
		})
	}

	tag.CssColor = newColor
	tag.Hidden = c.FormValue("hidden", "true") != "false"
	tag.ShortLine = c.FormValue("short_line")

	db.Save(&tag)

	return c.Render("partials/tag_wiki", fiber.Map{
		"Tag": tag,
	})
}

func GetTagWiki(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tag m.Tag
	db.First(&tag, c.Params("tag_id"))

	return c.Render("partials/tag_wiki", fiber.Map{
		"Tag": tag,
	})
}
