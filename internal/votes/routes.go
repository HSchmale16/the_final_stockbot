package votes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	vote := app.Group("/htmx/votes")
	vote.Get("/member/:memberId", GetVotesForMember)
	vote.Get("/:voteId", GetVote)
}

func GetVote(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	voteId := c.Params("voteId")

	var vote Vote
	db.Debug().Preload("VoteRecords").Preload("VoteRecords.Member").First(&vote, "id = ?", voteId)

	// Bin them by vote_status

	var voteStatus = map[string][]VoteRecord{}
	for _, record := range vote.VoteRecords {
		voteStatus[record.VoteStatus] = append(voteStatus[record.VoteStatus], record)
	}

	return c.Render("vote_table", fiber.Map{
		"Title":      "Vote",
		"Vote":       vote,
		"VoteStatus": voteStatus,
	})
}

func GetVotesForMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	memberId := c.Params("memberId")

	var results []struct {
		VoteType   string
		VoteStatus string
		Count      int
	}

	db.Debug().Model(&VoteRecord{}).
		InnerJoins("Vote").
		Select("vote_type, vote_status, count(*) as count").
		Where("member_id = ?", memberId).
		Group("vote_type, vote_status").
		Scan(&results)

	// Create a map to store the contingency table
	contingencyTable := make(map[string]map[string]int)

	// Iterate over the results and populate the contingency table
	for _, result := range results {
		voteType := result.VoteType
		voteStatus := result.VoteStatus
		count := result.Count

		// Check if the vote type exists in the contingency table
		if _, ok := contingencyTable[voteType]; !ok {
			contingencyTable[voteType] = make(map[string]int)
		}

		// Update the count for the vote status in the contingency table
		contingencyTable[voteType][voteStatus] = count
	}

	// Render the contingency table as a JSON response
	return c.JSON(fiber.Map{
		"ContingencyTable": contingencyTable,
	})
}
