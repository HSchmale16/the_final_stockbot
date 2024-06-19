package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/handlebars/v2"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed html_templates/*
var templates embed.FS

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
	engine.Load()

	for k, _ := range engine.Templates {
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

	// Middleware to pass db instance
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	app.Static("/static", "./static")

	// Setup the Routes
	app.Get("/", Index)
	app.Get("/tag/:tag_id", TagIndex)
	app.Post("/search", Search)

	app.Listen(":8080")
}

func Index(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var count int64
	db.Debug().Model(&GovtRssItemTag{}).Count(&count)

	// c.HTML(200, "index.html", gin.H{
	// 	"TagCount": count,
	// })

	return c.Render("index", fiber.Map{
		"Title":    "Hello, World!",
		"TagCount": count,
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

func Search(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var results []struct {
		TagId int64
		Name  string
		Count int64
	}

	db.Debug().Model(&GovtRssItemTag{}).
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

func main() {
	//DoBigApp()
	SetupServer()
}
