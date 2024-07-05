package app

import (
	"encoding/json"
	"html"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"golang.org/x/text/message"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetupServer() {
	db, err := SetupDB()
	if err != nil {
		panic(err)
	}

	//engine := handlebars.New("./html_templates", ".hbs")

	app := fiber.New(fiber.Config{
		Views: GetTemplateEngine(),
	})

	app.Use(helmet.New(helmet.Config{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'self'; frame-src 'self';",
	}))

	// Serve static files only on debug mode
	if os.Getenv("DEBUG") == "true" {
		app.Static("/static", "./static")
	}

	// Logging Request ID
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format: "${pid} ${latency} ${status} - ${method} ${path}?${queryParams}\n",
	}))

	// Middleware to pass db instance
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	CacheBustTimestamp := time.Now().Unix()

	IsDebug := os.Getenv("DEBUG") == "true"

	app.Use(func(c *fiber.Ctx) error {
		c.Bind(fiber.Map{
			"CacheBust": CacheBustTimestamp,
			"Title":     "DirtyCongress.com",
			"DEBUG":     IsDebug,
		})
		return c.Next()
	})
	app.Use(helmet.New())

	cacheMW := cache.New(cache.Config{})

	// Setup the Routes
	app.Get("/", Index)
	app.Get("/tags", TagList)
	app.Get("/tag/:tag_id", TagIndex)
	app.Get("/htmx/topic-search", TopicSearch)
	app.Get("/law/:law_id", LawView)
	app.Get("/law/:law_id/mods", LawView)
	app.Get("/laws", LawIndex)
	app.Get("/help", cacheMW, func(c *fiber.Ctx) error {
		return c.Render("help", fiber.Map{}, "layouts/main")
	})
	app.Get("/json/congress-network", CongressNetwork)
	app.Get("/congress-network", cacheMW, func(c *fiber.Ctx) error {
		return c.Render("congress_network", fiber.Map{
			"Title": "Congress Network Visualization",
		}, "layouts/main")
	})
	app.Get("/tos", cacheMW, TermsOfService)

	// HTMX End Point
	app.Use("/law/:law_id/tags", func(c *fiber.Ctx) error {
		db := c.Locals("db").(*gorm.DB)

		law_id := c.Params("law_id")

		var tags []struct {
			TagId int64
			Name  string
		}
		db.Raw("SELECT tag.id as tag_id, tag.name FROM tag JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id WHERE govt_rss_item_tag.govt_rss_item_id = ?", law_id).Scan(&tags)

		return c.Render("tag_search", fiber.Map{
			"Tags": tags,
		})
	})
	app.Get("/congress-members", cacheMW, CongressMemberList)
	app.Get("/congress-member/:bio_guide_id", ViewCongressMember)
	app.Get("/congress-member/:bio_guide_id/embed", EmbedCongressMember)
	app.Get("/htmx/congress_member/:bio_guide_id/finances", CongressMemberFinances)
	app.Get("/htmx/congress_member/:bio_guide_id/works_with", CongressMemberWorksWith)
	app.Get("/htmx/law/:law_id/related_laws", RelatedLaws)

	app.Listen(":8080")
}

func RelatedLaws(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var govtLaw GovtRssItem
	db.First(&govtLaw, c.Params("law_id"))

	title := govtLaw.Title
	// decode the html entities in title
	before, _, _ := strings.Cut(title, "(")

	x := strings.ReplaceAll(before, "&nbsp;", "%") + " %"

	var govtLaws []GovtRssItem
	db.Where("title LIKE ?", x).Where("ID != ?", govtLaw.ID).Limit(10).Find(&govtLaws)

	return c.Render("partials/law-list", fiber.Map{
		"Laws":     govtLaws,
		"SubTitle": "Related Laws",
	})
}

func TagList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tags []struct {
		ID    int64
		Name  string
		Count int64
	}
	db.Raw("SELECT tag.id, tag.name, COUNT(*) as count FROM tag JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id GROUP BY tag.id ORDER BY count DESC LIMIT 500").Scan(&tags)

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
	LIMIT := 7

	var laws []GovtRssItem
	// Pub date before
	x := db.Order("pub_date DESC").Limit(LIMIT) //.Find(&laws)

	if page != "missing" {
		x.Where("pub_date < ?", page).Find(&laws)
		// we don't use a layout here for htmx.
		// fuck if I get why I'm using htmx
		return c.Render("partials/law-list", fiber.Map{
			"Title":      "Most Recent Laws",
			"Laws":       laws,
			"EnableLoad": true,
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

	// metadata := ReadLawModsData(lawText.ModsXML)

	return c.Render("law_view", fiber.Map{
		"Title":      html.UnescapeString(law.Title),
		"Law":        law,
		"LawText":    lawText,
		"RenderMods": strings.HasSuffix(c.Path(), "/mods"),
		// "Metadata": metadata,
	}, "layouts/main")
}

func TermsOfService(c *fiber.Ctx) error {
	return c.Render("tos", fiber.Map{}, "layouts/main")
}

func CongressMemberList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var members []DB_CongressMember
	db.Find(&members)

	// Partition the members by being active
	// TODO: Expose more of the terms via the sql database
	var activeMembers []DB_CongressMember
	for _, member := range members {
		if member.IsActiveMember() {
			activeMembers = append(activeMembers, member)
		}
	}

	// Sort each by state
	sort.Slice(activeMembers, func(i, j int) bool {
		return activeMembers[i].State() < activeMembers[j].State()
	})

	return c.Render("congress_member_list", fiber.Map{
		"ActiveMembers": activeMembers,
	}, "layouts/main")
}

func ViewCongressMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.Preload("Sponsored").First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	return c.Render("congress_member_view", fiber.Map{
		"Member": member,
	}, "layouts/main")
}

func CongressMemberFinances(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	fecIds := member.CongressMemberInfo.Id.Fec
	var orgs []fecwrangling.CampaignCanidateLinkage

	db.Find(&orgs, "candidate_id IN ?", fecIds)

	return c.Render("partials/congress_member_finances", fiber.Map{
		"Orgs": orgs,
	})
}

func CongressMemberWorksWith(c *fiber.Ctx) error {
	// Do a simple graph search in both directions of degree 2 to see who they work with most often

	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	// Get the bills they sponsored
	var sponsored []CongressMemberSponsored
	db.Where("db_congress_member_bio_guide_id = ?", member.BioGuideId).Find(&sponsored)

	bills := make([]uint, len(sponsored))
	for i, bill := range sponsored {
		bills[i] = bill.GovtRssItemId
	}

	var sponsoredBy []string
	db.
		Model(&CongressMemberSponsored{}).
		Distinct("db_congress_member_bio_guide_id").
		Where("govt_rss_item_id IN ?", bills).
		Find(&sponsoredBy)

	// Get the members they work with
	var worksWith []DB_CongressMember
	db.Where("bio_guide_id IN ?", sponsoredBy).Find(&worksWith)

	parties := make(map[string]int)
	for _, member := range worksWith {
		parties[member.Party()]++
	}
	// dump to json
	data, _ := json.Marshal(parties)

	return c.Render("partials/congress_member_works_with", fiber.Map{
		"Member":    member,
		"WorksWith": worksWith,
		"Parties":   string(data),
	})
}

func EmbedCongressMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	return c.Render("embed/congress_member", fiber.Map{
		"Member": member,
		"Image":  "/static/img/muddy-" + string(member.Party()[0]) + ".jpg",
	})
}
