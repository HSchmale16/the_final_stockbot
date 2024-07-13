package app

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"gorm.io/gorm"
)

func SponsorsBillsWithPiChart(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	// Get the bills they sponsored
	var sponsored []CongressMemberSponsored
	db.Where("db_congress_member_bio_guide_id = ?", member.BioGuideId).Find(&sponsored)

	bills := make([]uint, len(sponsored))
	for i, bill := range sponsored {
		bills[i] = bill.GovtRssItemId
	}

	var sponsoredBy []string
	db.
		Model(&CongressMemberSponsored{}).
		Distinct("db_congress_member_bio_guide_id").
		Where("govt_rss_item_id IN ?", bills).
		Find(&sponsoredBy)

	// Get the members they work with
	var worksWith []DB_CongressMember
	db.Where("bio_guide_id IN ?", sponsoredBy).Find(&worksWith)

	parties := make(map[string]float64)
	for _, member := range worksWith {
		parties[member.Party()]++
	}

	c.Set("Content-Type", "image/svg+xml")

	b, err := renderPieChart(parties, member)
	if err != nil {
		return err
	}

	return c.Send(b)
}

func getColorForParty(party string) drawing.Color {
	switch string(party[0]) {
	case "R":
		return drawing.Color{R: 255, G: 0, B: 0, A: 255}
	case "D":
		return chart.ColorBlue
	case "I":
		return drawing.Color{R: 0xA0, G: 0x20, B: 0xF0, A: 255}
	default:
		return chart.ColorLightGray
	}
}

func renderPieChart(parties map[string]float64, member DB_CongressMember) ([]byte, error) {
	values := make([]chart.Value, 0, len(parties))
	for k, v := range parties {
		values = append(values, chart.Value{
			Value: v,
			Label: k,
			Style: chart.Style{
				FillColor: getColorForParty(k),
			},
		})
	}

	// Create a new pie chart
	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: values,
		//Title:  "Which Parties " + member.Name + " works with",
	}

	// Render the pie chart as a PNG image
	b := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.SVG, b)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
