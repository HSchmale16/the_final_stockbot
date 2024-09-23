package congress

import (
	"embed"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

type DB_CongressCommittee = m.DB_CongressCommittee
type DB_CommitteeMembership = m.DB_CommitteeMembership

//go:embed html_templates/*
var templateFS embed.FS

func init() {
	m.RegisterDebugFilePath("internal/congress/html_templates")
	m.RegisterEmbededFS(templateFS)
}

func SetupRoutes(app *fiber.App) {
	app.Get("/json/overlap/subcommittees/:thomas_id", CommitteeOverlap)
	app.Get("/committee_explorer", CommitteeExplorer)
	app.Get("/committee/:thomas_id", CommitteeView)
	app.Get("/committees", CommitteeList)
	app.Get("/bills", BillList)
	app.Get("/bill/:congress_number/:bill_type/:bill_number", BillView)
}

func BillList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var bills []Bill
	db.Find(&bills)

	return c.Render("bill_list", fiber.Map{
		"Title":       "Bill List",
		"Bills":       bills,
		"Description": "List of all bills in the US Congress, understand their scope and membership.",
	}, "layouts/main")
}

func BillView(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var bill Bill
	db.Debug().Preload("Actions", func(db *gorm.DB) *gorm.DB {
		return db.Order("bill_actions.action_time DESC")
	}).
		Preload("Actions.Committee").
		Preload("Actions.Vote").
		Preload("Actions.Vote.VoteRecords").
		Preload("Cosponsors.Member").
		First(&bill, "bill_type = ? AND congress_number = ? AND bill_number = ?", c.Params("bill_type"), c.Params("congress_number"), c.Params("bill_number"))

	return c.Render("bill_view", fiber.Map{
		"Title": bill.FormatTitle(),
		"Bill":  bill,
	}, "layouts/main")
}

func CommitteeExplorer(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var committees []DB_CongressCommittee
	db.Where("parent_committee_id IS NULL").Find(&committees)

	return c.Render("committee_explorer", fiber.Map{
		"Title":      "Committee Explorer",
		"Committees": committees,
	}, "layouts/main")
}

func CommitteeOverlap(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var committees DB_CongressCommittee
	db.Where("thomas_id = ?", c.Params("thomas_id")).Preload("Subcommittees.Memberships").Find(&committees)

	overlaps := computeOverlap(committees.Subcommittees)

	return c.JSON(overlaps)
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
