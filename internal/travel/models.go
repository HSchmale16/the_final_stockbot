package travel

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

/**
We are wrangling the travel disclosures of the congress critters.
*/

func init() {
	m.RegisterModels(&DB_TravelDisclosure{})
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

	// Where the travel disclosure came from
	Src string

	// A doc id to prevent duplicates. Comes from upstream
	DocId         string `gorm:"index"`
	FilerName     string
	MemberName    string
	Year          string `gorm:"index"`
	FilingType    string
	DepartureDate time.Time
	ReturnDate    time.Time
	Destination   string
	TravelSponsor string

	// These fields only appear on senate filings
	DateRecieved    time.Time
	TransactionDate time.Time

	// A url to download the original document from
	DocURL string
	// WHere it's stored on the local system
	Filepath string

	MemberId string `gorm:"index"`
	Member   m.DB_CongressMember
}

func (d DB_TravelDisclosure) TableName() string {
	return "travel_disclosures"
}

func LoadHouseXml(rc io.ReadCloser, db *gorm.DB) {
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

				if disclosure.State == "" {
					// Skip no state disclosure
					continue
				}

				// Check for duplicates by doc id
				var count int64
				db.Model(&DB_TravelDisclosure{}).Where("doc_id = ?", disclosure.DocId).Count(&count)
				if count > 0 {
					fmt.Println("Skipping duplicate", disclosure.DocId)
					continue
				}

				arr := strings.Split(disclosure.MemberName, ",")
				last := strings.Trim(arr[0], " .")
				first := strings.Trim(arr[1], " .")

				// split the first name and use the longer value
				splitFirst := strings.Split(first, " ")
				if len(splitFirst) > 1 {
					if len(splitFirst[0]) > len(splitFirst[1]) {
						first = splitFirst[0]
					} else {
						first = splitFirst[1]
					}
				}

				// Find the congress member
				var members []m.DB_CongressMember
				var member m.DB_CongressMember

				// Special case for Steve Watkins
				if first == "Steven" {
					first = "Steve"
				}

				rpStr := "REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(name, 'á', 'a'), 'ó', 'o'), 'ú', 'u'), 'é', 'e'), '’', ''''), '’', ''''), 'í', 'i')"

				// Really stupid answer to the problem of accented characters
				db.Debug().
					Where(rpStr+" ILIKE '%' || ? || '%'", last).
					Where(rpStr+" ILIKE '%' || ? || '%'", first).
					Find(&members)

				if len(members) == 1 {
					member = members[0]
				} else if len(members) > 1 {
					for _, m := range members {
						year, _ := strconv.Atoi(disclosure.Year)
						if m.CongressMemberInfo.ServedDuringYear(year) {
							member = m
						}
					}
				}

				if member.BioGuideId == "" {
					log.Fatal("Could not find member", disclosure.MemberName)
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

				// Create the disclosure
				db.Create(&DB_TravelDisclosure{
					Src:           "house",
					DocId:         disclosure.DocId,
					FilerName:     disclosure.FilerName,
					MemberName:    disclosure.MemberName,
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

func LoadSenateXml(rc io.ReadCloser, db *gorm.DB) {
	decoder := xml.NewDecoder(rc)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "dbo.filer" {
				var filer senateFilerXml
				decoder.DecodeElement(&filer, &se)

				loadSenateFiler(db, filer)

			}
		}
	}
}

func FuzzyFindSenator(db *gorm.DB, last, first, state string) (m.DB_CongressMember, error) {
	first = strings.Trim(first, ".")
	last = strings.Trim(last, ".")
	senator := last + ", " + first

	var member m.DB_CongressMember
	query1 := db.Debug().Model(&m.DB_CongressMember{}).
		Where("jsonb_path_query_first(congress_member_info, '$.terms[last].type')#>>'{}' = ?", "sen").
		Where("UNACCENT(UPPER(name)) ILIKE '%' || ? || '%'", last)
	if state != "" {
		query1 = query1.Where("jsonb_path_query_first(congress_member_info, '$.terms[last].state')#>>'{}' = ?", state)
	}

	var cnt int64
	query1.Count(&cnt)
	fmt.Println(cnt)
	if cnt > 1 {
		if first == "BOB" && last == "CASEY" {
			first = "ROBERT"
		}
		both := query1.Where("UNACCENT(UPPER(name)) ILIKE '%' || ? || '%'", first)
		both.Count(&cnt)
		fmt.Println(cnt)
		if cnt > 1 {
			// Print all records found
			var members []m.DB_CongressMember
			both.Find(&members)
			for _, m := range members {
				// Scan manually to see if we can find a year between them
				// We are only looking back at most 6 years for senators.
				if m.CongressMemberInfo.ServedDuringYear(2021) || m.CongressMemberInfo.ServedDuringYear(2023) {
					member = m
					break
				}
			}
		} else if cnt == 1 {
			both.First(&member)
		}
	} else if cnt == 1 {
		query1.First(&member)
	}

	if cnt == 0 {
		return m.DB_CongressMember{}, fmt.Errorf("no members found f='%s' l='%s' '%s'", first, last, senator)
	}

	if member.BioGuideId == "" {
		return m.DB_CongressMember{}, fmt.Errorf("could not find member %s", senator)
	}

	return member, nil
}

func loadSenateFiler(db *gorm.DB, filer senateFilerXml) {
	filerName := filer.FirstName + " " + filer.LastName

	// fmt.Println(filerName, senator)

	for _, office := range filer.Office {
		senator := office.OfficeName
		if senator == "BERRY, SONCERIA" {
			continue
		}
		if strings.Contains(senator, "SECRETARY FOR THE") {
			// skip non congress critters
			continue
		}

		last, first := m.SplitName(senator)
		member, err := FuzzyFindSenator(db, last, first, "")
		if err != nil {
			log.Fatal("Failed to fuzzy find a senator", err)
		}

		for _, doc := range office.Documents {
			// Check for duplicates by doc id
			var count int64
			db.Model(&DB_TravelDisclosure{}).Where("doc_id = ?", doc.Reports[0].DocURL).Count(&count)
			if count > 0 {
				log.Println("Skipping duplicate", doc.Reports[0].DocURL)
				continue
			}
			if len(doc.Reports) > 1 {
				log.Fatal("Multiple reports for a single document")
			}

			// Parse the dates
			departureDate, err := time.Parse("01/02/2006", doc.BeginTravelDate)
			if err != nil {
				log.Fatal("Error parsing departure date", doc.BeginTravelDate, err)
			}
			returnDate, err := time.Parse("01/02/2006", doc.EndTravelDate)
			if err != nil {
				log.Fatal("Error parsing return date", doc.EndTravelDate, err)
			}
			transactionDate, err := time.Parse("01/02/2006", doc.TransactionDate)
			if err != nil {
				log.Fatal("Error parsing transaction date", doc.TransactionDate, err)
			}
			recvDate, err := time.Parse("01/02/2006", doc.DateRecieved)
			if err != nil {
				log.Fatal("Error parsing received date", doc.DateRecieved, err)
			}

			x := DB_TravelDisclosure{
				Src:             "senate",
				FilerName:       filerName,
				MemberName:      senator,
				Year:            doc.Year,
				DocId:           doc.Reports[0].DocURL,
				DocURL:          doc.Reports[0].DocURL,
				FilingType:      doc.Reports[0].ReportTitle,
				DepartureDate:   departureDate,
				ReturnDate:      returnDate,
				TransactionDate: transactionDate,
				DateRecieved:    recvDate,
				MemberId:        member.BioGuideId,
			}

			db.Create(&x)
		}
	}

}

type senateFilerXml struct {
	FirstName string `xml:"FirstName,attr"`
	LastName  string `xml:"LastName,attr"`

	Office []senateOfficeXml `xml:"dbo.Office"`
}

type senateOfficeXml struct {
	OfficeName string `xml:"OfficeName,attr"`

	Documents []senateDocumentXml `xml:"dbo.Document"`
}

type senateDocumentXml struct {
	Year            string `xml:"ReportingYear,attr"`
	BeginTravelDate string `xml:"BeginTravelDate,attr"`
	EndTravelDate   string `xml:"EndTravelDate,attr"`
	DateRecieved    string `xml:"DateReceived,attr"`
	TransactionDate string `xml:"TransactionDate,attr"`
	Pages           string `xml:"Pages,attr"`

	Reports []senateReportXml `xml:"dbo.Reports"`
}

type senateReportXml struct {
	ReportTitle string `xml:"ReportTitle,attr"`
	DocURL      string `xml:"DocURL,attr"`
}
