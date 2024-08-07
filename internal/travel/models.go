package travel

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

/**
We are wrangling the travel disclosures of the congress critters.
*/

func init() {
	m.RegisterModels(&DB_TravelDisclosure{})
}

func SetupRoutes(app *fiber.App) {
	// Setup the routes
	app.Get("/htmx/congress-member/:id/travel", GetTravelDisclosures)
	app.Get("/htmx/recent-gift-travel", GetRecentGiftTravel)
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

	return c.Render("recent_gift_travel", fiber.Map{
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

type TravelDisclosure struct {
	DocId         string `xml:"DocID"`
	FilerName     string `xml:"FilerName"`
	MemberName    string `xml:"MemberName"`
	State         string `xml:"State"`
	Year          string `xml:"Year"`
	District      string `xml:"District"`
	Destination   string `xml:"Destination"`
	FilingType    string `xml:"FilingType"`
	DepartureDate string `xml:"DepartureDate"`
	ReturnDate    string `xml:"ReturnDate"`
	TravelSponsor string `xml:"TravelSponsor"`
}

type DB_TravelDisclosure struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	DocId         string
	FilerName     string
	Year          string
	FilingType    string
	DepartureDate time.Time
	ReturnDate    time.Time
	Destination   string
	TravelSponsor string

	MemberId string
	Member   m.DB_CongressMember
}

func LoadXml(rc io.ReadCloser, db *gorm.DB) {
	// Parse the XML file
	decoder := xml.NewDecoder(rc)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Travel" {
				var disclosure TravelDisclosure
				decoder.DecodeElement(&disclosure, &se)
				fmt.Println(disclosure)

				arr := strings.Split(disclosure.MemberName, ",")
				last := strings.Trim(arr[0], " ")
				first := strings.Trim(arr[1], " ")

				// Find the congress member
				var member m.DB_CongressMember
				db.
					Where("name LIKE '%' || ? || '%'", last).
					Where("name LIKE '%' || ? || '%'", first).
					First(&member)

				if member.BioGuideId == "" {
					fmt.Println("Could not find member", disclosure.MemberName)
					continue
				}

				// Parse the dates
				departureDate, err := time.Parse("1/2/2006", disclosure.DepartureDate)
				if err != nil {
					fmt.Println("Error parsing departure date", disclosure.DepartureDate, err)
					continue
				}

				returnDate, err := time.Parse("1/2/2006", disclosure.ReturnDate)
				if err != nil {
					fmt.Println("Error parsing return date", disclosure.ReturnDate)
					continue
				}

				// Check for duplicates by doc id
				var count int64
				db.Model(&DB_TravelDisclosure{}).Where("doc_id = ?", disclosure.DocId).Count(&count)
				if count > 0 {
					fmt.Println("Skipping duplicate", disclosure.DocId)
					continue
				}

				// Create the disclosure
				db.Create(&DB_TravelDisclosure{
					DocId:         disclosure.DocId,
					FilerName:     disclosure.FilerName,
					Year:          disclosure.Year,
					FilingType:    disclosure.FilingType,
					DepartureDate: departureDate,
					ReturnDate:    returnDate,
					Destination:   disclosure.Destination,
					TravelSponsor: disclosure.TravelSponsor,
					MemberId:      member.BioGuideId,
				})

			}
		}
	}
}
