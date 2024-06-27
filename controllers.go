package main

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"golang.org/x/text/message"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetupServer() {
	db, err := setupDB()
	if err != nil {
		panic(err)
	}

	//engine := handlebars.New("./html_templates", ".hbs")

	app := fiber.New(fiber.Config{
		Views: GetTemplateEngine(),
	})

	// Logging Request ID
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format: "${pid} ${latency} ${locals:requestid} ${status} - ${method} ${path}\n",
	}))

	// Serve static files only on debug mode
	if os.Getenv("DEBUG") == "true" {

		app.Static("/static", "./static")
	}

	// Middleware to pass db instance
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	app.Static("/static", "./static")

	// Setup the Routes
	app.Get("/", Index)
	app.Get("/tags", TagList)
	app.Get("/tag/:tag_id", TagIndex)
	app.Get("/htmx/topic-search", TopicSearch)
	app.Get("/law/:law_id", LawView)
	app.Get("/laws", LawIndex)
	app.Get("/help", func(c *fiber.Ctx) error {
		return c.Render("help", fiber.Map{}, "layouts/main")
	})
	app.Get("/congress-network", CongressNetwork)

	app.Listen(":8080")
}

func TagList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tags []Tag
	db.Raw("SELECT tag.id, tag.name, COUNT(*) as count FROM tag JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id GROUP BY tag.id ORDER BY count DESC").Limit(200).Scan(&tags)

	return c.Render("tag_list", fiber.Map{
		"Tags": tags,
	}, "layouts/main")
}

func Index(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var articleTags, totalTags, totalLaws int64
	db.Model(&GovtRssItemTag{}).Count(&articleTags)
	db.Model(&Tag{}).Count(&totalTags)
	db.Model(&GovtRssItem{}).Count(&totalLaws)

	p := message.NewPrinter(message.MatchLanguage("en"))

	return c.Render("index", fiber.Map{
		"Title":       "Congress Magnifying Glass",
		"TotalTopics": p.Sprintf("%d", articleTags),
		"TotalTags":   p.Sprintf("%d", totalTags),
		"TotalLaws":   p.Sprintf("%d", totalLaws),
	}, "layouts/main")
}

func LawIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var laws []GovtRssItem
	// Pub date before
	x := db.Debug().Order("pub_date DESC").Limit(50) //.Find(&laws)
	if c.FormValue("before") != "" {
		x = x.Where("pub_date <= ?", c.FormValue("before"))
	}
	x.Find(&laws)

	return c.Render("law_index", fiber.Map{
		"Title": "Most Recent Laws",
		"Laws":  laws,
	}, "layouts/main")
}

func TagIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tag Tag
	db.First(&tag, c.Params("tag_id"))

	var items []GovtRssItem
	db.Model(&GovtRssItem{}).
		Joins("JOIN govt_rss_item_tag ON govt_rss_item_tag.govt_rss_item_id = govt_rss_item.id").
		Where("govt_rss_item_tag.tag_id = ?", tag.ID).
		Order("pub_date DESC").
		Limit(100).
		Preload(clause.Associations).
		Find(&items)

	return c.Render("tag_index", fiber.Map{
		"Tag":   tag,
		"Items": items,
	}, "layouts/main")
}

func TopicSearch(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var results []struct {
		TagId int64
		Name  string
		Count int64
	}

	db.Model(&GovtRssItemTag{}).
		Select("tag_id, Name, COUNT(*) as count").
		Joins("JOIN tag ON tag.id = tag_id").
		Joins("Join govt_rss_item ON govt_rss_item.id = govt_rss_item_id").
		Where("LOWER(tag.name) LIKE LOWER(?)", "%"+strings.ToLower(c.FormValue("search"))+"%").
		Where("govt_rss_item.pub_date > ?", "2023").
		Group("tag_id").
		Order("COUNT(*) DESC").
		Limit(500).
		Scan(&results)

	var minCount, maxCount int64
	if len(results) > 0 {
		minCount = results[0].Count
		maxCount = results[0].Count
		for _, result := range results {
			if result.Count < minCount {
				minCount = result.Count
			}
			if result.Count > maxCount {
				maxCount = result.Count
			}
		}
	}

	db.Create(&SearchQuery{
		Query:      c.FormValue("search"),
		NumResults: len(results),
	})

	return c.Render("tag_search", fiber.Map{
		"Tags":     results,
		"MinCount": minCount,
		"MaxCount": maxCount,
	})
}

func LawView(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var law GovtRssItem
	db.Preload(clause.Associations).Find(&law, c.Params("law_id"))

	var lawText GovtLawText
	db.First(&lawText, "govt_rss_item_id = ?", law.ID)

	metadata := ReadLawModsData(lawText.ModsXML)

	return c.Render("law_view", fiber.Map{
		"Title":    html.UnescapeString(law.Title),
		"Law":      law,
		"LawText":  lawText,
		"Metadata": metadata,
	}, "layouts/main")
}

func CongressNetwork(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var results []struct {
		RssId   int64
		ModsXML string
		PubDate time.Time
	}
	db.Raw("SELECT govt_rss_item.id as rss_id, govt_rss_item.pub_date, govt_law_text.mods_xml FROM govt_rss_item JOIN govt_law_text ON govt_law_text.govt_rss_item_id = govt_rss_item.id").Scan(&results)

	// Create a bigraph of all the congress critters who work together
	type Edge struct {
		Source CongressMember
		Target CongressMember
	}

	edges := make(map[Edge]int)

	for _, result := range results {
		mods := ReadLawModsData(result.ModsXML)
		if len(mods.CongressMembers) == 0 {
			continue
		}
		sponser := mods.CongressMembers[0]
		for _, member := range mods.CongressMembers[1:] {
			edge := Edge{
				Source: sponser,
				Target: member,
			}
			edges[edge]++
		}
	}

	type s struct {
		Edge
		Count int
	}

	edges_array := make([]s, 0, len(edges))

	for edge, count := range edges {
		edges_array = append(edges_array, s{
			Edge:  edge,
			Count: count,
		})
	}
	fmt.Println(len(edges_array))

	return c.Render("congress_network", fiber.Map{
		"Edges": edges_array,
	}, "layouts/main")
}
