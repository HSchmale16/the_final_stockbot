package lobbying

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/lobbying/:year", RenderLobbyingPage)
	app.Get("/lobbying/breakdown/:year/:type", RenderBreakdownPage)
}

var ContributionType = map[string]string{
	"feca": "FECA",
	"he":   "Honorary Expenses",
	"me":   "Meeting Expenses",
	"ple":  "Presidential Library Expenses",
	"pic":  "Presidential Inaugural Committee",
}

func RenderLobbyingPage(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	year := c.Params("year")

	type Row struct {
		ContributionTypeDisplay string
		ContributionType        string `gorm:"column:contribution_type"`
		Count                   int
		Amount                  float64
		AmountStr               string
	}

	var lobbyingContributions []Row

	year2, _ := strconv.Atoi(year)
	db.Raw("SELECT contribution_type, SUM(CAST(amount AS float)) Amount, Count(*) as Count FROM lobbyist_contributions WHERE filing_year = ? GROUP BY contribution_type", year2).Scan(&lobbyingContributions)
	p := message.NewPrinter(language.English)

	// iterate and update the contribution type with the diplay version
	for i, row := range lobbyingContributions {
		lobbyingContributions[i].ContributionTypeDisplay = ContributionType[row.ContributionType]
		lobbyingContributions[i].AmountStr = p.Sprintf("%.2f", row.Amount)
	}

	fmt.Println(year, lobbyingContributions)

	return c.Render("lobbying", fiber.Map{
		"Title": "Lobbying Spending for " + c.Params("year"),
		"Year":  c.Params("year"),
		"Data":  lobbyingContributions,
	}, "layouts/main")
}

func RenderBreakdownPage(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	year := c.Params("year")
	contributionType := c.Params("type")

	type Row struct {
		ContributorName string `gorm:"column:registrant_name"`
		PayeeName       string `gorm:"column:payee_name"`
		Count           int
		Amount          float64
		AmountStr       string
	}

	var lobbyingContributions []Row

	year2, _ := strconv.Atoi(year)
	db.Raw("SELECT registrant_name, payee_name, SUM(CAST(amount AS float)) Amount, Count(*) as Count FROM lobbyist_contributions WHERE filing_year = ? AND contribution_type = ? GROUP BY registrant_name, payee_name ORDER BY Amount DESC LIMIT 30", year2, contributionType).Scan(&lobbyingContributions)
	p := message.NewPrinter(language.English)

	// iterate and update the contribution type with the diplay version
	for i, row := range lobbyingContributions {
		lobbyingContributions[i].AmountStr = p.Sprintf("%.2f", row.Amount)
	}

	return c.Render("lobbying_breakdown", fiber.Map{
		"Title": "Lobbying Spending for " + c.Params("year"),
		"Year":  c.Params("year"),
		"Type":  ContributionType[contributionType],
		"Data":  lobbyingContributions,
	}, "layouts/main")
}
