package congress

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	. "github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/committee/:thomas_id", CommitteeView)
	app.Get("/committees", CommitteeList)
}

func CommitteeList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	committeeType := c.Query("type")

	var committees []DB_CongressCommittee
	x := db.Preload("Subcommittees").
		Where("parent_committee_id IS NULL")

	if committeeType != "" {
		x = x.Where("type = ?", committeeType)
	}

	dbc := x.Find(&committees)
	if dbc.Error != nil {
		fmt.Println("Error", dbc.Error)
		return c.Status(404).SendString("404 Not Found")
	}

	return c.Render("committee_list", fiber.Map{
		"Title":         "Committee List",
		"Description":   "List of all committees in the US Congress, understand their scope and membership.",
		"Committees":    committees,
		"CommitteeType": committeeType,
	}, "layouts/main")
}

func CommitteeView(c *fiber.Ctx) error {
	var committee DB_CongressCommittee
	db := c.Locals("db").(*gorm.DB)

	dbc := db.
		Preload("Memberships.CongressMember").
		Preload("Subcommittees.Memberships.CongressMember").
		First(&committee, "thomas_id = ?", c.Params("thomas_id"))

	if committee.Name == "" {
		return c.Status(404).SendString("404 Not Found")
	}
	if dbc.Error != nil {
		fmt.Println("Error", dbc.Error)
		return c.Status(404).SendString("404 Not Found")
	}

	// The preloading is being stupid if we try to sort up there.
	// So we'll just define a function over it and live with it.
	// There's no more than 20ish members so NBD
	committee.SortMembers()
	for i := range committee.Subcommittees {
		committee.Subcommittees[i].SortMembers()
	}

	if committee.ParentCommitteeId != nil {
		db.First(&committee.ParentCommittee, "thomas_id = ?", committee.ParentCommitteeId)
	}
	db.Limit(5).Order("pub_date desc").Model(&committee).Association("GovtRssItems").Find(&committee.GovtRssItems)

	return c.Render("committee_view", fiber.Map{
		"Title":       committee.Name,
		"Description": "View the " + committee.Name + " in the US Congress, understand their scope and membership.",
		"Committee":   committee,
	}, "layouts/main")
}
