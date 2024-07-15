package congress

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	. "github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/committee/:thomas_id", CommitteeView)
}

func CommitteeView(c *fiber.Ctx) error {
	var committee DB_CongressCommittee
	db := c.Locals("db").(*gorm.DB)

	dbc := db.Debug().
		Preload("Memberships.CongressMember").
		Preload("Subcommittees.Memberships.CongressMember").
		First(&committee, "thomas_id = ?", c.Params("thomas_id"))
	if dbc.Error != nil {
		fmt.Println("Error", dbc.Error)
		return c.Status(404).SendString("404 Not Found")
	}

	fmt.Println("Committee", len(committee.Memberships))

	return c.Render("committee_view", fiber.Map{
		"Title":     "Committee",
		"Committee": committee,
	}, "layouts/main")
}
