package main

import (
	"embed"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/handlebars/v2"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed html_templates/*
var templates embed.FS

//go:embed all:static
var embedDirStatic embed.FS

func SetupServer() {
	db, err := setupDB()
	if err != nil {
		panic(err)
	}

	//engine := handlebars.New("./html_templates", ".hbs")
	subFS, err := fs.Sub(templates, "html_templates")
	if err != nil {
		panic(err)
	}
	engine := handlebars.NewFileSystem(http.FS(subFS), ".hbs")

	for k := range engine.Templates {
		fmt.Println(k)
	}

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Logging Request ID
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format: "${pid} ${latency} ${locals:requestid} ${status} - ${method} ${path}\n",
	}))

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "static",
		Browse:     true,
	}))

	// Middleware to pass db instance
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	app.Static("/static", "./static")

	// Setup the Routes
	app.Get("/", Index)
	app.Get("/tag/:tag_id", TagIndex)
	app.Get("/htmx/topic-search", TopicSearch)
	app.Get("/law/:law_id", LawView)
	app.Get("/laws", LawIndex)

	app.Listen(":8080")
}

func Index(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var articleTags, totalTags int64
	db.Model(&GovtRssItemTag{}).Count(&articleTags)
	db.Model(&Tag{}).Count(&totalTags)

	// c.HTML(200, "index.html", gin.H{
	// 	"TagCount": count,
	// })

	return c.Render("index", fiber.Map{
		"Title":       "Hello, World!",
		"TotalTopics": articleTags,
		"TotalTags":   totalTags,
	}, "layouts/main")
}

func LawIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var laws []GovtRssItem
	db.Order("pub_date DESC").Limit(10).First(&laws)

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
		Limit(300).
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

	// c.HTML(200, "tag_search.html", gin.H{
	// 	"Tags":     results,
	// 	"MinCount": minCount,
	// 	"MaxCount": maxCount,
	// })

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

	return c.Render("law_view", fiber.Map{
		"Title":   html.UnescapeString(law.Title),
		"Law":     law,
		"LawText": lawText,
	}, "layouts/main")
}

func main() {
	//FederalRegisterParser()
	go DoBigApp()
	SetupServer()
}
