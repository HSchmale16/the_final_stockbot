package travel

import (
	"log"
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	// Setup the routes
	app.Get("/htmx/congress-member/:id/travel", GetTravelDisclosures)
	app.Get("/htmx/recent-gift-travel", GetRecentGiftTravel)
	app.Get("/travel-by-destination/:destination", GetTravelByDestination)
	app.Get("/htmx/top-destinations", GetTopDestinations)
	app.Get("/htmx/travel/committee/:committee", GetTravelByCommittee)
}

func GetTravelByCommittee(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	committee, err := url.PathUnescape(c.Params("committee"))
	if err != nil {
		return c.Status(400).SendString("Invalid committee")
	}

	var committeeDB m.DB_CongressCommittee
	db.Find(&committeeDB, "thomas_id = ?", committee)

	// Get the travel disclosures
	var disclosures []DB_TravelDisclosure
	errDb := db.
		Joins("JOIN db_committee_memberships ON db_committee_memberships.db_congress_member_bio_guide_id = travel_disclosures.member_id").
		Where("db_committee_memberships.db_congress_committee_thomas_id = ?", committeeDB.ThomasId).
		Order("departure_date DESC").
		Preload("Member").
		Limit(75).
		Find(&disclosures)

	if errDb.Error != nil {
		log.Default().Print(errDb.Error)
		return c.Status(400).SendString("Invalid things")
	}

	return c.Render("htmx/travel_list", fiber.Map{
		"Title":      "Gifted Travel for " + committeeDB.Name,
		"PageTitle":  "Gifted Travel for " + committeeDB.Name,
		"GiftTravel": disclosures,
	})
}

func GetTopDestinations(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// Get limit param
	limitStr, _ := strconv.Atoi(c.Query("limit"))
	limit := min(150, max(15, limitStr))

	var topDestinations []struct {
		Destination string
		Count       int
	}
	x := db.Table("travel_disclosures").
		Select("destination, count(destination) as count").
		Where("destination != ''").
		Group("destination").
		Order("count DESC").
		Limit(limit).
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

	var interfaceDisclosures []interface{}
	for _, d := range disclosures {
		interfaceDisclosures = append(interfaceDisclosures, d)
	}

	PartyBreakdown := m.MakeSponsorshipMap(disclosures, func(i DB_TravelDisclosure) string {
		return i.Member.Party()
	})

	return c.Render("htmx/gift_travel_to_dest", fiber.Map{
		"Title":          "Gifted Travel to " + destination,
		"PageTitle":      "Gifted Travel to " + destination,
		"GiftTravel":     disclosures,
		"PartyBreakdown": PartyBreakdown,
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
