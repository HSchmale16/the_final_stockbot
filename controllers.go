package main

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format: "${pid} ${latency} ${status} - ${method} ${path}?${queryParams}\n",
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
	app.Get("/json/congress-network", CongressNetwork)
	app.Get("/congress-network", CongressNetworkLayout)
	app.Get("/tos", TermsOfService)

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

	var recentLaws []GovtRssItem = make([]GovtRssItem, 0, 10)
	db.Order("pub_date DESC").Limit(10).Find(&recentLaws)

	p := message.NewPrinter(message.MatchLanguage("en"))

	return c.Render("index", fiber.Map{
		"Title":       "DirtyCongress.com",
		"TotalTopics": p.Sprintf("%d", articleTags),
		"TotalTags":   p.Sprintf("%d", totalTags),
		"TotalLaws":   p.Sprintf("%d", totalLaws),
		"Laws":        recentLaws,
	}, "layouts/main")
}

func LawIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// convert to int minus 1
	page := c.Query("page", "missing")
	LIMIT := 20

	var laws []GovtRssItem
	// Pub date before
	x := db.Order("pub_date DESC").Limit(LIMIT) //.Find(&laws)

	if page != "missing" {
		x.Where("pub_date < ?", page).Find(&laws)
		// we don't use a layout here for htmx.
		// fuck if I get why I'm using htmx
		return c.Render("partials/law-list", fiber.Map{
			"Title": "Most Recent Laws",
			"Laws":  laws,
		})
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

	chamber := c.FormValue("chamber")

	var results []struct {
		RssId   int64
		ModsXML string
		PubDate time.Time
	}
	db.Raw("SELECT govt_rss_item.id as rss_id, govt_rss_item.pub_date, govt_law_text.mods_xml FROM govt_rss_item JOIN govt_law_text ON govt_law_text.govt_rss_item_id = govt_rss_item.id").Scan(&results)

	// Create a bigraph of all the congress critters who work together
	type Edge struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	edges := make(map[Edge]int)
	nodes := make(map[string]CongressMember)

	for _, result := range results {
		mods := ReadLawModsData(result.ModsXML)
		if len(mods.CongressMembers) == 0 || mods.CongressMembers[0].Chamber != chamber {
			continue
		}
		sponser := mods.CongressMembers[0]
		nodes[sponser.BioGuideId] = sponser
		for _, member := range mods.CongressMembers[1:] {
			edge := Edge{
				Source: sponser.BioGuideId,
				Target: member.BioGuideId,
			}
			nodes[member.BioGuideId] = member
			edges[edge] += 2

			// reverse the edge
			edge = Edge{
				Source: member.BioGuideId,
				Target: sponser.BioGuideId,
			}
			edges[edge]++
		}
	}

	type s struct {
		Edge
		Value int `json:"value"`
	}

	edges_array := make([]s, 0, len(edges))

	for edge, count := range edges {
		edges_array = append(edges_array, s{
			Edge:  edge,
			Value: count,
		})
	}
	fmt.Println(len(edges_array))

	node_values := make([]CongressMember, 0, len(nodes))
	for _, v := range nodes {
		node_values = append(node_values, v)
	}

	return c.JSON(fiber.Map{
		"nodes": node_values,
		"edges": edges_array,
	})
}

func CongressNetworkLayout(c *fiber.Ctx) error {
	return c.Render("congress_network", fiber.Map{
		"Title": "Congress Network Visualization",
	}, "layouts/main")
}

func TermsOfService(c *fiber.Ctx) error {
	return c.Render("tos", fiber.Map{}, "layouts/main")
}
