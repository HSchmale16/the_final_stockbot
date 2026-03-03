package travel

import (
	"embed"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

//go:embed html_templates/*
var templateFS embed.FS

func init() {
	m.RegisterDebugFilePath("internal/travel/html_templates")
	m.RegisterEmbededFS(templateFS)
}

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

	app.Get("/travel/calendar/:year/:month", GetTravelCalendar)
	app.Get("/json/travel/calendar/:year/:month", GetTripsInYearMonth)
	app.Get("/json/travel/calendar2", GetCalendar2)
	app.Get("/json/travel/calendar_data/:year/:month", GetTravelCalendarDataAsJson)

	app.Get("/gifted-travel2", GetGiftedTravel2)
	app.Get("/htmx/gifted-travel-rows/:date/:sponsor", GetGiftedTravelRows)
	app.Get("/travel/calendar", RedirectToPriorMonthCalendar)
	app.Get("/htmx/travel-gauge-cluster", TravelGaugeCluster)
}

func TravelGaugeCluster(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var totalTravelers int64
	db.Model(&DB_TravelDisclosure{}).
		Where("filing_type != ?", "Ammendment").
		Distinct("member_id").
		Count(&totalTravelers)

	var totalTrips int64
	db.Model(&DB_TravelDisclosure{}).
		Where("filing_type != ?", "Ammendment").
		Count(&totalTrips)

	var totalDays int64
	db.Model(&DB_TravelDisclosure{}).
		Where("filing_type != ?", "Ammendment").
		Select("COALESCE(SUM(return_date::date - departure_date::date), 0)").
		Scan(&totalDays)

	return c.Render("htmx/travel_gauge_cluster", fiber.Map{
		"TotalTravelers": totalTravelers,
		"TotalTrips":     totalTrips,
		"TotalDays":      totalDays,
	})
}

func RedirectToPriorMonthCalendar(c *fiber.Ctx) error {
	now := time.Now()
	// Redirect to the previous month as there is never data for the current month
	previousMonth := now.AddDate(0, -1, 0)
	year := previousMonth.Year()
	month := int(previousMonth.Month())
	return c.Redirect(fmt.Sprintf("/travel/calendar/%d/%d", year, month))
}

func GetTravelCalendarDataAsJson(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	year, _ := strconv.Atoi(c.Params("year"))
	month, _ := strconv.Atoi(c.Params("month"))

	weeks := GetTravelCalendarData(year, month, db)
	return c.JSON(weeks)
}

type TravelGroup struct {
	DepartureDate time.Time
	TravelSponsor string
	Count         int
}

func GetGiftedTravelRows(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	date := c.Params("date")
	sponsor, err := url.QueryUnescape(c.Params("sponsor"))
	if err != nil {
		return c.Status(400).SendString("Invalid sponsor")
	}

	fmt.Println(sponsor, date)
	var disclosures []DB_TravelDisclosure
	db.
		Where("departure_date = ?", date).
		Where("travel_sponsor = ?", sponsor).
		Preload("Member").
		Order("member_id DESC").
		Find(&disclosures)

	return c.Render("htmx/gifted_travel_rows", fiber.Map{
		"GiftTravel": disclosures,
	})
}

func GetGiftedTravel2(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var travelGroups []TravelGroup
	db.Model(&DB_TravelDisclosure{}).
		Select("departure_date, travel_sponsor, COUNT(*) as count").
		Group("departure_date, travel_sponsor").
		Order("departure_date DESC").
		Limit(50).
		Find(&travelGroups)

	return c.Render("gifted_travel2", fiber.Map{
		"Title":        "Grouped Gifted Travel",
		"TravelGroups": travelGroups,
	}, "layouts/main")
}

func GetCalendar2(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	targetDate := c.Query("targetDate")
	if targetDate == "" {
		return c.Status(400).SendString("Invalid targetDate")
	}

	var disclosures []DB_TravelDisclosure
	db.
		// Where("TO_DATE(?, 'YYYY-MM-DD') BETWEEN SYMMETRIC DATE(departure_date) AND DATE(return_date)", targetDate).
		Where("departure_date::date <= ?::date", targetDate).
		Where("return_date::date >= ?::date", targetDate).
		Preload("Member").
		Order("destination ASC").
		Find(&disclosures)

	return c.JSON(disclosures)
}

func GetTripsInYearMonth(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	year := c.Params("year")
	month := c.Params("month")

	dateStr := fmt.Sprintf("%s-%s-01", year, month)

	var trips []DB_TravelDisclosure
	db.
		Preload("Member").
		Where("?::date in (date_trunc('month', departure_date)::date, date_trunc('month', return_date)::date)", dateStr).
		Find(&trips)

	return c.JSON(trips)
}

func GetTravelCalendar(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// make ints
	year := c.Params("year")
	month := c.Params("month")

	// Get year
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return c.Status(400).SendString("Invalid year")
	}

	// Get month name
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		return c.Status(400).SendString("Invalid month")
	}
	monthName := time.Month(monthInt).String()

	return c.Render("travel_calendar", fiber.Map{
		"Title":     "Gifted Travel Calendar - " + monthName + " " + year,
		"PrevMonth": time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0).Format("2006/01"),
		"NextMonth": time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Format("2006/01"),
		"weeks":     GetTravelCalendarData(yearInt, monthInt, db),
		"yearInt":   yearInt,
		"monthInt":  monthInt,
	}, "layouts/main")
}

func GetTravelHomepage(c *fiber.Ctx) error {
	year := c.Query("year")
	if year == "" {
		year = fmt.Sprint(time.Now().Year())
	}
	return c.Render("travel_homepage", fiber.Map{
		"Title":   "Gifted Travel Disclosures",
		"OgImage": "static/img/og-travel-plots.png",
		"Year":    year,
	}, "layouts/main")
}

func GetMostTravelTable(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// Check the year is a number
	_, err := strconv.Atoi(c.Params("year"))
	if err != nil {
		return c.Status(400).SendString("Invalid year")
	}
	year := c.Params("year")

	var data []struct {
		MemberId string
		Member   m.DB_CongressMember
		Count    int
	}

	db.Model(DB_TravelDisclosure{}).
		Where("filing_type != ?", "Ammendment").
		Where("year = ?", year).
		Group("member_id").
		Select("member_id, Count(*) as count").
		Order("count DESC").
		Limit(15).
		Scan(&data)

	// N+1 query problem here. TODO: FIX
	for i, d := range data {
		db.Find(&data[i].Member, "bio_guide_id = ?", d.MemberId)
	}

	return c.Render("htmx/most_travel_table", fiber.Map{
		"MostTravel":   data,
		"SelectedYear": year,
		"Years":        []int{2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025, 2026},
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
		Where("filing_type in ?", []string{"Original", "Member Reimbursed Travel"}).
		Joins("Inner Join congress_member cm ON cm.bio_guide_id = member_id").
		Group("year, jsonb_path_query_first(congress_member_info, '$.terms[last].party')").
		Select("year, jsonb_path_query_first(congress_member_info, '$.terms[last].party')#>>'{}' as party, sum(EXTRACT(DAY FROM return_date - departure_date)) as count").
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

	db.Table(DB_TravelDisclosure{}.TableName()).
		Where("filing_type in ?", []string{"Original", "Member Reimbursed Travel"}).
		Joins("Inner Join congress_member cm ON cm.bio_guide_id = member_id").
		Group("year, jsonb_path_query_first(congress_member_info, '$.terms[last].party')").
		Select("year, jsonb_path_query_first(congress_member_info, '$.terms[last].party')#>>'{}' as party, Count(*) as count").
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

// GetTopDestinations returns the top travel destinations for congress members.
//
// Parameters (via query string):
//   - limit: The maximum number of destinations to return (range: 15-150, default: 15).
//   - since: A starting date in YYYY-MM-DD format. Mutually exclusive with 'year'.
//   - year:  A specific year to filter by. Mutually exclusive with 'since'.
//
// Returns:
//   - 400 Bad Request if both 'since' and 'year' are provided.
//   - 400 Bad Request if 'since' date format is invalid.
//   - 500 Internal Server Error if the database connection is missing.
func GetTopDestinations(c *fiber.Ctx) error {
	since := c.Query("since")
	year := c.Query("year")

	if since != "" && year != "" {
		return c.Status(400).SendString("Parameters 'since' and 'year' are mutually exclusive")
	}

	var sinceDate time.Time
	if since != "" {
		var err error
		sinceDate, err = time.Parse("2006-01-02", since)
		if err != nil {
			return c.Status(400).SendString("Invalid 'since' date format. Use YYYY-MM-DD")
		}
	}

	dbIn := c.Locals("db")
	db, ok := dbIn.(*gorm.DB)
	if !ok || db == nil {
		return c.Status(500).SendString("Database not found")
	}

	// Get limit param
	limitStr, _ := strconv.Atoi(c.Query("limit"))
	limit := min(150, max(15, limitStr))

	var topDestinations []struct {
		Destination string
		Count       int
	}

	query := db.Table("travel_disclosures").
		Select("destination, count(destination) as count").
		Where("destination != ''")

	if db.Config != nil && db.Config.Logger != nil {
		query = query.Debug()
	}

	if since != "" {
		query = query.Where("departure_date >= ?", sinceDate)
	} else if year != "" {
		query = query.Where("year = ?", year)
	} else {
		a := time.Now().Year()
		b := a - 1
		aStr := fmt.Sprint(a)
		bStr := fmt.Sprint(b)
		query = query.Where("year in (?, ?)", aStr, bStr)
	}

	x := query.Group("destination").
		Order("count DESC").
		Limit(limit).
		Scan(&topDestinations)

	if x.Error != nil {
		log.Default().Print(x.Error)
		return x.Error
	}

	return c.Render("htmx/top_destinations", fiber.Map{
		"TopDestinations": topDestinations,
		"Since":           since,
		"Year":            year,
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
		"Destination":    destination,
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
