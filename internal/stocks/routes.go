package stocks

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/congress-member/:id/finances", RenderFinances)
}

func RenderFinances(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var finDisclosureDocuments []FinDisclosureDocument
	db.Model(&FinDisclosureDocument{}).Where("member_id = ?", c.Params("id")).Order("filing_date DESC").Find(&finDisclosureDocuments)

	return c.Render("partials/congress_member_finances", fiber.Map{
		"Title":       "Finances",
		"Disclosures": finDisclosureDocuments,
	})
}
