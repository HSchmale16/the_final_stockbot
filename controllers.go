package main

import (
	"html"
	"html/template"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetupServer() {
	db, err := setupDB()
	if err != nil {
		panic(err)
	}

	// Setup the gin server
	r := gin.Default()

	funcMap := template.FuncMap{
		"uri_decode": func(s string) string {
			res, err := url.QueryUnescape(s)
			if err != nil {
				return s
			}
			return res
		},
		"unescape": func(s string) string {
			return html.UnescapeString(s)
		},
	}

	r.SetFuncMap(funcMap)

	// Load the HTML templates
	r.LoadHTMLGlob("html_templates/*")

	// Setup the database
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	r.Static("/static", "./static")

	// Setup the Routes
	r.GET("/", Index)
	r.GET("/tag/:tag_id", TagIndex)
	r.POST("/search", Search)
	r.Run(":8080")
}

func Index(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Debug().Model(&GovtRssItemTag{}).Count(&count)

	c.HTML(200, "index.html", gin.H{
		"TotalTags": count,
	})
}

func TagIndex(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var tag Tag
	db.Debug().First(&tag, c.Param("tag_id"))

	var items []GovtRssItem
	db.Debug().Model(&GovtRssItem{}).
		Joins("JOIN govt_rss_item_tag ON govt_rss_item_tag.govt_rss_item_id = govt_rss_item.id").
		Where("govt_rss_item_tag.tag_id = ?", tag.ID).
		Order("pub_date DESC").
		Limit(100).
		Preload(clause.Associations).
		Find(&items)

	c.HTML(200, "tag_index.html", gin.H{
		"Tag":   tag,
		"Items": items,
	})
}

func Search(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var results []struct {
		TagId int64
		Name  string
		Count int64
	}

	db.Debug().Model(&GovtRssItemTag{}).
		Select("tag_id, Name, COUNT(*) as count").
		Joins("JOIN tag ON tag.id = tag_id").
		Joins("Join govt_rss_item ON govt_rss_item.id = govt_rss_item_id").
		Where("LOWER(tag.name) LIKE LOWER(?)", "%"+strings.ToLower(c.PostForm("search"))+"%").
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

	c.HTML(200, "tag_search.html", gin.H{
		"Tags":     results,
		"MinCount": minCount,
		"MaxCount": maxCount,
	})
}

func main() {
	//DoBigApp()
	SetupServer()
}
