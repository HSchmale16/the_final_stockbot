package lobbying

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// Maximum number of records able to returned by the SQL interface
	MAX_SQL_RECORDS = 350
)

var YearsLoaded = []string{"2018", "2019", "2020", "2021", "2022", "2023"}
var LobbyingTypes = []string{"feca", "he", "me", "ple", "pic"}
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

	app.Get("/lobbying-sql", LobbyingSQLView)
	app.Post("/lobbying-sql", ExecLobbyistSQL)
}

var ContributionType = map[string]string{
	"feca": "FECA",
	"he":   "Honorary Expenses",
	"me":   "Meeting Expenses",
	"ple":  "Presidential Library Expenses",
	"pic":  "Presidential Inaugural Committee",
}

func LobbyingSQLView(c *fiber.Ctx) error {
	return c.Render("lobbying_sql", fiber.Map{
		"Title":       "Query Facts about Lobbying in Congress - DirtyCongress.com",
		"Description": "Query the database for information about lobbying in Congress. Understand more about the types of contributions made by lobbyists.",
	}, "layouts/main")
}

type row struct {
	Values []string
}

type resultSet struct {
	Columns []string
	Rows    []row
}

func IsUnsafeSQLForLobbyists(sql string) bool {
	// Joins cause it to run out of memory for some reason.
	return strings.Contains(sql, "JOIN") ||
		strings.Contains(sql, "UNION") ||
		strings.Contains(sql, "DELETE") ||
		strings.Contains(sql, "DROP") ||
		strings.Contains(sql, "TRUNCATE") ||
		strings.Contains(sql, "ALTER") ||
		strings.Contains(sql, "CREATE") ||
		strings.Contains(sql, "INSERT") ||
		strings.Contains(sql, "UPDATE") ||
		strings.Contains(sql, "REPLACE") ||
		strings.Contains(sql, "INTO") ||
		strings.Contains(sql, "SET") ||
		strings.Contains(sql, "duckdb_settings()")
}

func ExecLobbyistSQL(c *fiber.Ctx) error {
	sql := c.FormValue("sql")
	sql = strings.TrimSpace(sql)

	if IsUnsafeSQLForLobbyists(sql) {
		LogAnalytics(sql, fmt.Errorf("invalid sql detected by filter"), 0, c)
		return c.SendString("Invalid SQL: Stop Trying to Hack Me :(")
	}

	var x resultSet

	i := 0
	rows, err := LobbyingDBInstance.DB.Query(sql)
	if err != nil {
		fmt.Println(err)
	} else {
		x.Columns, _ = rows.Columns()
		defer rows.Close()

		for rows.Next() {
			var r row

			// holy shit this is absolutely miserable to get implemented.
			// I can't believe how difficult this was to get working
			values := make([]interface{}, len(x.Columns))
			for i := range values {
				values[i] = new(interface{})
			}

			err := rows.Scan(values...)
			if err != nil {
				fmt.Println(err)
				break
			}

			for _, val := range values {
				r.Values = append(r.Values, fmt.Sprintf("%v", *val.(*interface{})))
			}

			x.Rows = append(x.Rows, r)

			i++
			if i > MAX_SQL_RECORDS {
				break
			}
		}
	}

	LogAnalytics(sql, err, i, c)

	return c.Render("table", fiber.Map{
		"Error": err,
		"Rows":  x,
	})
}

func shittyString(err error) *string {
	if err != nil {
		x := err.Error()
		return &x
	}
	return nil
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

	year2, err := strconv.Atoi(year)
	if err != nil {
		return c.Status(404).SendString("Invalid year")
	}

	rows, err := LobbyingDBInstance.DB.Query("SELECT contribution_type, Count(*) as Count, SUM(CAST(amount AS float)) Amount FROM contributions_etl WHERE filing_year = ? GROUP BY contribution_type ORDER BY Amount DESC", year2)

	if err != nil {
		return err
	}
	defer rows.Close()

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
		"Title":       "Lobbying Spending for " + c.Params("year"),
		"Year":        c.Params("year"),
		"Description": "Lobbying Spending by type for " + c.Params("year") + " in Congress. Understand more about the types of contributions made by lobbyists.",
		"Data":        lobbyingContributions,
		"Years":       YearsLoaded,
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
		"Description": "Lobbying Spending by type for " + c.Params("year") + " in Congress. Understand more about the types of contributions made by lobbyists.",
		"Year":        c.Params("year"),
		"TypeDisplay": ContributionType[contributionType],
		"Type":        contributionType,
		"Data":        lobbyingContributions,
		"Years":       YearsLoaded,
	}, "layouts/main")
}
