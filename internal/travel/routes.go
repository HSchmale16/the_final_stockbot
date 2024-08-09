package travel

import (
	"log"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	// Setup the routes
	app.Get("/htmx/congress-member/:id/travel", GetTravelDisclosures)
	app.Get("/htmx/recent-gift-travel", GetRecentGiftTravel)
	app.Get("/travel-by-destination/:destination", GetTravelByDestination)
	app.Get("/htmx/top-destinations", GetTopDestinations)
}

func GetTopDestinations(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var topDestinations []struct {
		Destination string
		Count       int
	}
	x := db.Table("travel_disclosures").
		Select("destination, count(destination) as count").
		Where("destination != ''").
		Group("destination").
		Order("count DESC").
		Limit(15).
		Scan(&topDestinations)

	if x.Error != nil {
		log.Default().Print(x.Error)
		return x.Error
	}

	return c.Render("htmx/top_destinations", fiber.Map{
		"TopDestinations": topDestinations,
	})
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

	return c.Render("htmx/recent_gift_travel", fiber.Map{
		"Title":      "Gifted Travel to " + destination,
		"PageTitle":  "Gifted Travel to " + destination,
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

	return c.Render("htmx/recent_gift_travel", fiber.Map{
		"GiftTravel": disclosures,
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
