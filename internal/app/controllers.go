package app

import (
	"html"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/hschmale16/the_final_stockbot/internal/congress"
	"github.com/hschmale16/the_final_stockbot/internal/faq"
	"github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	. "github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/stocks"
	"github.com/hschmale16/the_final_stockbot/internal/travel"
	"golang.org/x/text/message"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DOMAIN = "https://www.dirtycongress.com"
)

func SetupServer() {
	db, err := SetupDB()
	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		Views: GetTemplateEngine(),
		// Required if I want to get the ip address of actual requests.
		// Powered by NGINX
		ProxyHeader: fiber.HeaderXForwardedFor,
	})

	app.Use(helmet.New(helmet.Config{
		HSTSPreloadEnabled: true,
		HSTSMaxAge:         15768000,
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
			"CacheBust":   CacheBustTimestamp,
			"Title":       "Dirty Congress",
			"DEBUG":       IsDebug,
			"Description": "DirtyCongress.com provides a searchable database of bills and congress members with advanced visualizations of lobbying and other contributions to congress.",
			"Url":         DOMAIN + c.OriginalURL(),
			"Url2":        c.OriginalURL(),
		})
		return c.Next()
	})
	app.Use(helmet.New())

	if !IsDebug {
		app.Use(cache.New(cache.Config{
			KeyGenerator: func(c *fiber.Ctx) string {
				return utils.CopyString(c.Path()) + utils.CopyString(string(c.Context().URI().QueryString()))
			},
			Expiration: 5 * time.Minute,
		}))
	}

	// Setup the Routes
	app.Get("/", Index)
	app.Get("/tags", TagList)
	app.Get("/tag/:tag_id", TagIndex)

	app.Get("/htmx/topic-search", TopicSearch)
	app.Get("/htmx/tag-datalist", TagDataList)
	app.Get("/law/:law_id", LawView)
	app.Get("/law/:law_id/mods", LawView)
	app.Get("/laws", LawIndex)
	app.Get("/json/congress-network", CongressNetwork)
	app.Get("/congress-network", func(c *fiber.Ctx) error {
		return c.Render("congress_network", fiber.Map{
			"Title": "Congress Network Visualization",
		}, "layouts/main")
	})
	app.Get("/tos", TermsOfService)

	// HTMX End Point
	app.Use("/law/:law_id/tags", func(c *fiber.Ctx) error {
		db := c.Locals("db").(*gorm.DB)

		law_id := c.Params("law_id")

		var tags []struct {
			TagId    int64
			Name     string
			CssColor string
		}
		db.Raw("SELECT tag.id as tag_id, tag.name, tag.css_color FROM tag JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id WHERE govt_rss_item_tag.govt_rss_item_id = ?", law_id).Scan(&tags)

		return c.Render("tag_search", fiber.Map{
			"Tags": tags,
		})
	})
	app.Get("/congress-members", CongressMemberList)
	app.Get("/congress-member/:bio_guide_id", ViewCongressMember)
	app.Get("/congress-member/:bio_guide_id/embed", EmbedCongressMember)
	app.Get("/congress-member/:bio_guide_id/sponsors-bills-with-pi-chart", SponsorsBillsWithPiChart)
	app.Get("/htmx/congress_member/:bio_guide_id/finances", CongressMemberFinances)
	app.Get("/htmx/congress_member/:bio_guide_id/works_with", CongressMemberWorksWith)
	app.Get("/htmx/law/:law_id/related_laws", RelatedLaws)

	// Helper to double check the sqlite pragmas
	app.Get("/meta/sqlite_info", func(c *fiber.Ctx) error {
		db := c.Locals("db").(*gorm.DB)

		var data []struct {
			CompileOptions string
		}

		db.Raw("Pragma compile_options").Scan(&data)

		return c.JSON(data)
	})

	faq.SetupRoutes(app)
	// lobbying.SetupRoutes(app)
	congress.SetupRoutes(app)
	stocks.SetupRoutes(app)
	travel.SetupRoutes(app)

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
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
	db.Raw("SELECT tag.id, tag.name, COUNT(*) as count FROM tag JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id GROUP BY tag.id ORDER BY count DESC LIMIT 1000").Scan(&tags)

	return c.Render("tag_list", fiber.Map{
		"Tags": tags,
	}, "layouts/main")
}

func TagDataList(c *fiber.Ctx) error {
	// HERE
	var tags []Tag

	db := c.Locals("db").(*gorm.DB)
	db.Debug().
		Where("name LIKE ?", "%"+c.FormValue("search")+"%").
		Joins("JOIN govt_rss_item_tag ON govt_rss_item_tag.tag_id = tag.id").
		Joins("Join congress_member_sponsored ON congress_member_sponsored.govt_rss_item_id = govt_rss_item_tag.govt_rss_item_id").
		Group("tag.id").
		Order("Count(*) DESC").
		Having("COUNT(*) > 1").
		Limit(10).
		Find(&tags)

	db.Create(&SearchQuery{
		Query:      c.FormValue("search"),
		NumResults: len(tags),
		IpAddr:     c.IP(),
		UserAgent:  c.Get("User-Agent"),
	})

	return c.Render("htmx/tag_datalist", fiber.Map{
		"Tags": tags,
	})
}

func Index(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var articleTags, totalTags, totalLaws int64
	db.Model(&GovtRssItemTag{}).Count(&articleTags)
	db.Model(&Tag{}).Count(&totalTags)
	db.Model(&GovtRssItem{}).Count(&totalLaws)

	var Approprations, Senate, Public, House, HouseRes, SenateRes []GovtRssItem

	billLimit := 7

	db.Preload("Sponsors").Order("pub_date DESC").Joins("JOIN govt_rss_item_tag ON govt_rss_item_tag.govt_rss_item_id = govt_rss_item.id").Limit(billLimit).Find(&Approprations, "tag_id = ?", 377)
	db.Preload("Sponsors").Order("pub_date DESC").Limit(billLimit).Find(&House, "title LIKE ?", "H.R.%")
	db.Preload("Sponsors").Order("pub_date DESC").Limit(billLimit).Find(&Senate, "title LIKE ?", "S. %")
	db.Preload("Sponsors").Order("pub_date DESC").Limit(billLimit).Find(&Public, "title LIKE ?", "Public Law %")

	db.Preload("Sponsors").Order("pub_date DESC").Limit(billLimit).Find(&HouseRes, "title LIKE ?", "H. Res.%")
	db.Preload("Sponsors").Order("pub_date DESC").Limit(billLimit).Find(&SenateRes, "title LIKE ?", "S. Res.%")

	// var recentLaws []GovtRssItem = make([]GovtRssItem, 0, 10)
	// db.Preload("Sponsors").Order("pub_date DESC").Limit(10).Find(&recentLaws)

	p := message.NewPrinter(message.MatchLanguage("en"))

	return c.Render("index", fiber.Map{
		"Title":         "Dirty Congress - Explore the Laws and Connections within Congress",
		"TotalTopics":   p.Sprintf("%d", articleTags),
		"TotalTags":     p.Sprintf("%d", totalTags),
		"TotalLaws":     p.Sprintf("%d", totalLaws),
		"Approprations": Approprations,
		"House":         House,
		"Senate":        Senate,
		"PublicLaws":    Public,
		"HouseRes":      HouseRes,
		"SenateRes":     SenateRes,
		"SearchValue":   c.FormValue("search"),
	}, "layouts/main")
}

func LawIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	page := c.Query("page", "missing")
	LIMIT := 7

	var laws []GovtRssItem
	// Pub date before
	x := db.Preload("Sponsors").Order("pub_date DESC").Limit(LIMIT)
	lawType := c.Query("type")

	if lawType != "" {
		x.Where("title LIKE ?", lawType+"%")
	}

	if page != "missing" {
		x.Where("pub_date < ?", page).Find(&laws)
		// we don't use a layout here for htmx.
		return c.Render("partials/law-list", fiber.Map{
			"Title":      "Most Recent " + GetLawTypeDisplay(lawType),
			"Laws":       laws,
			"EnableLoad": true,
			"LawType":    lawType,
		})
	}

	x.Find(&laws)

	return c.Render("law_index", fiber.Map{
		"Title":       "Most Recent " + GetLawTypeDisplay(lawType),
		"Description": "Understand the bills currently being debated in congress",
		"Laws":        laws,
		"LawType":     lawType,
	}, "layouts/main")
}

func TagIndex(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tag Tag
	db.First(&tag, c.Params("tag_id"))

	var items []GovtRssItem
	db.Debug().Model(&GovtRssItem{}).
		Preload("Sponsors").
		Joins("JOIN govt_rss_item_tag ON govt_rss_item_tag.govt_rss_item_id = govt_rss_item.id").
		Where("govt_rss_item_tag.tag_id = ?", tag.ID).
		Order("pub_date DESC").
		Limit(100).
		Find(&items)

	return c.Render("tag_index", fiber.Map{
		"Title":       "View Bills Tagged With " + tag.Name,
		"Description": "View bills tagged with " + tag.Name + " --- " + tag.ShortLine,
		"Tag":         tag,
		"Items":       items,
	}, "layouts/main")
}

func TopicSearch(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var results []struct {
		TagId     int64
		Name      string
		CssColor  string
		ShortLine string
		Count     int64
	}

	db.Model(&GovtRssItemTag{}).
		Select("tag_id, Name, css_color as CssColor, short_line as ShortLine, COUNT(*) as count").
		Joins("JOIN tag ON tag.id = tag_id").
		Joins("Join govt_rss_item ON govt_rss_item.id = govt_rss_item_id").
		Where("LOWER(tag.name) LIKE LOWER(?)", "%"+strings.ToLower(c.FormValue("search"))+"%").
		Where("govt_rss_item.pub_date > ?", "2023").
		Group("tag_id").
		Order("COUNT(*) DESC").
		Limit(250).
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

	// Try Searching Using the FTS Title Table
	var ftsResults []struct {
		RowId   int64 `gorm:"column:rowid"`
		Title   string
		PubDate string
	}

	search := c.FormValue("search")
	search += "*"
	db.Select("rowid, title, pub_date").Table("fts_law_title").Where("fts_law_title MATCH ?", search).Order("pub_date DESC").Limit(5).Scan(&ftsResults)

	db.Create(&SearchQuery{
		Query:      c.FormValue("search"),
		NumResults: len(results),
		IpAddr:     c.IP(),
		UserAgent:  c.Get("User-Agent"),
	})

	return c.Render("tag_search", fiber.Map{
		"Tags":     results,
		"FtsLaws":  ftsResults,
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

	determineFirstTitle := func(x string) string {
		// split based on the last instance of of a hyphen
		// this is because the title is in the format of "H.R. 1234 - Title"
		// and we want to get the "H.R. 1234" part
		split := strings.LastIndex(x, " - ")
		if split != -1 {
			return x[:split]
		}
		return x
	}

	return c.Render("law_view", fiber.Map{
		"Title":       html.UnescapeString(law.Title),
		"Description": "View the sponsors of this " + determineFirstTitle(law.Title) + " and the actual primary source text",
		"Law":         law,
		"LawText":     lawText,
		"RenderMods":  strings.HasSuffix(c.Path(), "/mods"),
	}, "layouts/main")
}

func TermsOfService(c *fiber.Ctx) error {
	return c.Render("tos", fiber.Map{
		"Title": "Terms of Service * Privacy Policy",
	}, "layouts/main")
}

func CongressMemberList(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var members []DB_CongressMember
	state := c.Query("state")

	if state != "" {
		db.Where("json_extract(congress_member_info, '$.terms[#-1].state') = ?", state).Find(&members)
	} else {
		db.Find(&members)
	}

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

	title := "Current Congress Members"
	if state != "" {
		title = "Current Congress Members from " + state
	}

	return c.Render("congress_member_list", fiber.Map{
		"ActiveMembers": activeMembers,
		"Description":   "A list of the current congress members",
		"Title":         title,
	}, "layouts/main")
}

func ViewCongressMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.
		Preload("Sponsored.Sponsors").
		Preload("Committees.Committee").
		First(&member, DB_CongressMember{
			BioGuideId: c.Params("bio_guide_id"),
		})

	return c.Render("congress_member_view", fiber.Map{
		"Title":       member.Name,
		"Description": "Understand more about how " + member.Name + " is connected to other congress members and the bills they have sponsored",
		"Member":      member,
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

	// Get the members that sponsored the same bills
	var sponsoredBy []struct {
		DB_CongressMember
		Count int
	}
	db.
		Table("congress_member").
		Joins("JOIN congress_member_sponsored ON db_congress_member_bio_guide_id = bio_guide_id").
		Where("bio_guide_id != ?", member.BioGuideId).
		Where("govt_rss_item_id IN ?", bills).
		Group("db_congress_member_bio_guide_id").
		Select("congress_member.*, COUNT(*) as Count").
		Order("Count DESC").
		Limit(30).
		//Having("COUNT(*) > 1").
		Scan(&sponsoredBy)

	return c.Render("partials/congress_member_works_with", fiber.Map{
		"Member":    member,
		"WorksWith": sponsoredBy,
	})
}

func EmbedCongressMember(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var member DB_CongressMember
	db.First(&member, DB_CongressMember{
		BioGuideId: c.Params("bio_guide_id"),
	})

	var numBillsSponsored int64
	db.Model(&CongressMemberSponsored{}).Where("db_congress_member_bio_guide_id = ?", member.BioGuideId).Count(&numBillsSponsored)

	return c.Render("embed/congress_member", fiber.Map{
		"Member":            member,
		"Image":             "/static/img/muddy-" + string(member.Party()[0]) + ".jpg",
		"NumBillsSponsored": numBillsSponsored,
	})
}
