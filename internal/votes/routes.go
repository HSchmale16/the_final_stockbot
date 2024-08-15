package votes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	vote := app.Group("/htmx/votes")
	vote.Get("/:memberId", GetVotesForMember)
}

func GetVotesForMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	memberId := c.Params("memberId")

	var results []struct {
		VoteType string
		Count    int
	}

	db.Model(&Vote{}).
		Select("vote_type, count(*) as count").
		Where("member_id = ?", memberId).
		Group("vote_type").
		Scan(&results)

	return c.Render("votes_members", fiber.Map{
		"Votes": results,
	})
}
