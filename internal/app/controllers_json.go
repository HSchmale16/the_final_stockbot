package app

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/hschmale16/the_final_stockbot/internal/m"
)

//go:embed sql/congress_network.sql
var congress_network_sql string

//go:embed sql/congress_network_tag.sql
var congress_network_tag_sql string

func CongressNetwork(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	edges, nodes, err := GetGraphNodes(db, c.Query("chamber"), c.Query("tag_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"edges": edges,
	})
}

func CongressNetworkHierarchy(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	edges, nodes, err := GetGraphNodes(db, c.Query("chamber"), c.Query("tag_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	BuildAStupidGraph(nodes, edges)

	return nil
}

func BuildAStupidGraph(nodes []CM_GraphNode, edges []CM_Edge) {

}

func GetGraphNodes(db *gorm.DB, chamber, tagId string) ([]CM_Edge, []CM_GraphNode, error) {
	var edges []CM_Edge

	if tagId != "" {
		log.Print("Using tag_id")
		tag_id_num, err := strconv.Atoi(tagId)
		if err != nil {
			return nil, nil, fmt.Errorf("tag_id must be an integer")
		}
		db.Raw(congress_network_tag_sql, chamber, tag_id_num).Scan(&edges)

		// log usage of special congress network views
		db.Create(&m.TagUse{
			TagId: uint(tag_id_num),
			// right now I don't care to log them at the moment.
			// IpAddr:    c.IP(),
			// UserAgent: c.Get("User-Agent"),
			UseType: "cn", // congress network usage
		})
	} else {
		db.Raw(congress_network_sql, chamber).Scan(&edges)
	}

	// Distinctify the people in source and target
	var node_names = make(map[string]bool)
	for _, edge := range edges {
		node_names[edge.Source] = true
		node_names[edge.Target] = true
	}
	keys := make([]string, 0, len(node_names))
	for k := range node_names {
		keys = append(keys, k)
	}

	// Select all the congress people mentioned in node_names keys
	var congress_people []struct {
		m.DB_CongressMember
		Count int
	}
	db.Table("congress_member").
		Select("congress_member.*", "(SELECT COUNT(*) FROM congress_member_sponsored WHERE db_congress_member_bio_guide_id = congress_member.bio_guide_id) as count").
		Where("bio_guide_id IN ?", keys).Find(&congress_people)

	var nodes = make([]CM_GraphNode, len(congress_people))

	for i, person := range congress_people {
		nodes[i].BioGuideId = person.BioGuideId
		nodes[i].Name = person.Name
		nodes[i].State = person.CongressMemberInfo.Terms[0].State
		nodes[i].Party = person.CongressMemberInfo.Terms[0].Party
		nodes[i].Count = person.Count
		nodes[i].RenderName = person.Name + " (" + person.State() + "-" + person.Party() + ")"
	}

	return edges, nodes, nil
}

type CM_Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

type CM_GraphNode struct {
	BioGuideId string
	RenderName string
	Name       string
	State      string
	Party      string
	Count      int
}
