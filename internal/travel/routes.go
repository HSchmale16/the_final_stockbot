package travel

import (
	"log"
	"net/url"
	"sort"
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
	app.Get("/travel", GetTravelHomepage)
	app.Get("/htmx/travel/who-travels-most/:year", GetMostTravelTable)
	app.Get("/json/travel-by-party", GetTravelByParty)
	app.Get("/json/days-traveled-by-party", GetDaysGiftedTravelByParty)
}

func GetTravelHomepage(c *fiber.Ctx) error {
	return c.Render("travel_homepage", fiber.Map{
		"Title":   "Gifted Travel Disclosures",
		"OgImage": "static/img/og-travel-plots.png",
	}, "layouts/main")
}

func GetMostTravelTable(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// Get the year
	year, err := strconv.Atoi(c.Params("year"))
	if err != nil {
		return c.Status(400).SendString("Invalid year")
	}

	var data []struct {
		DB_TravelDisclosure
		Count int
	}

	db.Debug().Model(&DB_TravelDisclosure{}).
		Where("year = ?", year).
		Joins("Member").
		Group("member_id").
		Select("member_id, Count(*) as count").
		Order("count DESC").
		Limit(15).
		Scan(&data)

	return c.Render("htmx/most_travel_table", fiber.Map{
		"MostTravel":   data,
		"SelectedYear": year,
		"Years":        []int{2018, 2019, 2020, 2021, 2022, 2023, 2024},
	})
}

func GetDaysGiftedTravelByParty(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var data []struct {
		Year  int    `json:"year"`
		Party string `json:"party"`
		Count int    `json:"count"`
	}

	db.Model(&DB_TravelDisclosure{}).
		Joins("Inner Join congress_member cm ON cm.bio_guide_id = member_id").
		Group("year, json_extract(congress_member_info, '$.terms[#-1].party')").
		Select("year, json_extract(congress_member_info, '$.terms[#-1].party') as party, sum(julianday(return_date) - julianday(departure_date)) as count").
		Scan(&data)

	result := make(map[int]map[string]int)

	for _, d := range data {
		year := d.Year
		party := d.Party
		count := d.Count

		if _, ok := result[year]; !ok {
			result[year] = make(map[string]int)
		}

		result[year][party] = count
	}

	result2 := make([]map[string]interface{}, 0)
	for year, parties := range result {
		partyMap := make(map[string]interface{})
		partyMap["year"] = year
		for party, count := range parties {
			partyMap[party] = count
		}
		result2 = append(result2, partyMap)
	}

	// sort by year
	sort.Slice(result2, func(i, j int) bool {
		return result2[i]["year"].(int) < result2[j]["year"].(int)
	})

	return c.JSON(result2)
}

func GetTravelByParty(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var data []struct {
		Year  int    `json:"year"`
		Party string `json:"party"`
		Count int    `json:"count"`
	}

	db.Model(&DB_TravelDisclosure{}).
		Joins("Inner Join congress_member cm ON cm.bio_guide_id = member_id").
		Group("year, json_extract(congress_member_info, '$.terms[#-1].party')").
		Select("year, json_extract(congress_member_info, '$.terms[#-1].party') as party, Count(*) as count").
		Scan(&data)

	result := make(map[int]map[string]int)

	for _, d := range data {
		year := d.Year
		party := d.Party
		count := d.Count

		if _, ok := result[year]; !ok {
			result[year] = make(map[string]int)
		}

		result[year][party] = count
	}

	result2 := make([]map[string]interface{}, 0)
	for year, parties := range result {
		partyMap := make(map[string]interface{})
		partyMap["year"] = year
		for party, count := range parties {
			partyMap[party] = count
		}
		result2 = append(result2, partyMap)
	}

	// sort by year
	sort.Slice(result2, func(i, j int) bool {
		return result2[i]["year"].(int) < result2[j]["year"].(int)
	})

	return c.JSON(result2)
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
