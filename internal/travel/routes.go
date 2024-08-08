package travel

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	// Setup the routes
	app.Get("/htmx/congress-member/:id/travel", GetTravelDisclosures)
	app.Get("/htmx/recent-gift-travel", GetRecentGiftTravel)
	app.Get("/travel-by-destination/:destination", GetTravelByDestination)
}

func GetTravelByDestination(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	destination, err := url.PathUnescape(c.Params("destination"))
	if err != nil {
		return c.Status(400).SendString("Invalid destination")
	}
	// Get the travel disclosures
	var disclosures []DB_TravelDisclosure
	db.
		Where("destination = ?", destination).
		Order("departure_date DESC").
		Preload("Member").
		Find(&disclosures)

	return c.Render("recent_gift_travel", fiber.Map{
		"Title":      "Gifted Travel to " + destination,
		"GiftTravel": disclosures,
	}, "layouts/main")
}

func GetRecentGiftTravel(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// Get the travel disclosures
	var disclosures []DB_TravelDisclosure
	db.
		Order("departure_date DESC").
		Preload("Member").
		Limit(10).
		Find(&disclosures)

	var topDestinations []struct {
		Destination string
		Count       int
	}
	db.Table("db_travel_disclosures").
		Select("destination, count(destination) as count").
		Group("destination").
		Order("count DESC").
		Limit(15).
		Scan(&topDestinations)

	return c.Render("recent_gift_travel", fiber.Map{
		"GiftTravel":      disclosures,
		"TopDestinations": topDestinations,
	})
}

func GetTravelDisclosures(c *fiber.Ctx) error {
	// Get the member id
	memberId := c.Params("id")
	db := c.Locals("db").(*gorm.DB)

	// Get the travel disclosures
	var disclosures []DB_TravelDisclosure
	db.
		Where("member_id = ?", memberId).
		Order("departure_date DESC").
		Find(&disclosures)

	return c.Render("travel_disclosures", fiber.Map{
		"GiftTravel": disclosures,
	})
}
