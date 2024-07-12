package lobbying

import (
	_ "embed"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var YearsLoaded = []string{"2018", "2019", "2020", "2021", "2022", "2023"}
var p = message.NewPrinter(language.English)

func SetupRoutes(app *fiber.App) {
	app.Get("/lobbying/:year", RenderLobbyingYearPage)
	app.Get("/lobbying/breakdown/:year/:type", RenderBreakdownPage)
	app.Get("/lobbying", func(c *fiber.Ctx) error {
		return c.Render("lobbying", fiber.Map{
			"Title": "Lobbyist Contributions Year Index",
			"Years": YearsLoaded,
		}, "layouts/main")
	})
}

var ContributionType = map[string]string{
	"feca": "FECA",
	"he":   "Honorary Expenses",
	"me":   "Meeting Expenses",
	"ple":  "Presidential Library Expenses",
	"pic":  "Presidential Inaugural Committee",
}

func RenderLobbyingYearPage(c *fiber.Ctx) error {
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

	rows, err := LobbyingDBInstance.DB.Query("SELECT contribution_type, Count(*) as Count, SUM(CAST(amount AS float)) Amount FROM contributions_etl WHERE filing_year = ? GROUP BY contribution_type ORDER BY Amount DESC", year2)
	if err != nil {
		return err
	}

	for rows.Next() {
		x := Row{}
		rows.Scan(&x.ContributionType, &x.Count, &x.Amount)
		lobbyingContributions = append(lobbyingContributions, x)
	}

	// iterate and update the contribution type with the diplay version
	for i, row := range lobbyingContributions {
		lobbyingContributions[i].ContributionTypeDisplay = ContributionType[row.ContributionType]
		lobbyingContributions[i].AmountStr = p.Sprintf("%.2f", row.Amount)
	}

	return c.Render("lobbying_types", fiber.Map{
		"Title": "Lobbying Spending for " + c.Params("year"),
		"Year":  c.Params("year"),
		"Data":  lobbyingContributions,
		"Years": YearsLoaded,
	}, "layouts/main")
}

//go:embed sql/contribution_breakdown.sql
var contributionSQL string

func RenderBreakdownPage(c *fiber.Ctx) error {
	year := c.Params("year")
	contributionType := c.Params("type")

	type Row struct {
		RegistrantName string `gorm:"column:registrant_name"`
		PayeeName      string `gorm:"column:payee_name"`
		HonoreeName    string
		Count          int
		Amount         float64
		AmountStr      string
	}

	var lobbyingContributions = make([]Row, 0, 50)

	year2, _ := strconv.Atoi(year)

	rows, err := LobbyingDBInstance.DB.Query(contributionSQL, year2, contributionType)
	if err != nil {
		return err
	}

	for rows.Next() {
		x := Row{}
		rows.Scan(&x.RegistrantName, &x.PayeeName, &x.HonoreeName, &x.Amount, &x.Count)
		lobbyingContributions = append(lobbyingContributions, x)
	}

	// iterate and update the contribution type with the diplay version
	for i, row := range lobbyingContributions {
		lobbyingContributions[i].AmountStr = p.Sprintf("%.2f", row.Amount)
	}

	return c.Render("lobbying_breakdown", fiber.Map{
		"Title":       "Lobbying Spending for " + c.Params("year"),
		"Year":        c.Params("year"),
		"TypeDisplay": ContributionType[contributionType],
		"Type":        contributionType,
		"Data":        lobbyingContributions,
		"Years":       YearsLoaded,
	}, "layouts/main")
}
