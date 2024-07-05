package app

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

//go:embed sql/congress_network.sql
var congress_network_sql string

func CongressNetwork(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	chamber := c.FormValue("chamber")

	var edges []struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Value  int    `json:"value"`
	}

	db.Raw(congress_network_sql, chamber).Scan(&edges)

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
		DB_CongressMember
		Count int
	}
	db.Table("congress_member").
		Select("congress_member.*", "(SELECT COUNT(*) FROM congress_member_sponsored WHERE db_congress_member_bio_guide_id = congress_member.bio_guide_id) as count").
		Where("bio_guide_id IN ?", keys).Find(&congress_people)

	type Node struct {
		BioGuideId string
		Name       string
		State      string
		Party      string
		Count      int
	}

	var nodes = make([]Node, len(congress_people))

	for i, person := range congress_people {
		nodes[i].BioGuideId = person.BioGuideId
		nodes[i].Name = person.Name
		nodes[i].State = person.CongressMemberInfo.Terms[0].State
		nodes[i].Party = person.CongressMemberInfo.Terms[0].Party
		nodes[i].Count = person.Count
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"edges": edges,
	})
}
