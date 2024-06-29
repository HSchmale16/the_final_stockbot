package main

import (
	_ "embed"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

//go:embed congress_network.sql
var congress_network_sql string

func CongressNetwork(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	chamber := c.FormValue("chamber")

	var edges []struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Value  int    `json:"value"`
	}

	db.Debug().Raw(congress_network_sql, chamber).Scan(&edges)

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
	var congress_people []DB_CongressMember
	db.Debug().Where("bio_guide_id IN ?", keys).Find(&congress_people)

	fmt.Println(len(congress_people), len(keys))

	type Node struct {
		BioGuideId string
		Name       string
		State      string
		Party      string
	}

	var nodes = make([]Node, len(congress_people))

	for i, person := range congress_people {
		nodes[i].BioGuideId = person.BioGuideId
		nodes[i].Name = person.Name
		nodes[i].State = person.CongressMemberInfo.Terms[0].State
		nodes[i].Party = person.CongressMemberInfo.Terms[0].Party
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"edges": edges,
	})
}
