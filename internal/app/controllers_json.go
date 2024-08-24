package app

import (
	_ "embed"
	"fmt"
	"log"
	"math"
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

	SetGroupsViaConnectedComponents(nodes, edges)
	for i, node := range nodes {
		nodes[i].R = math.Sqrt(float64(node.Count)) + 3.0
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"edges": edges,
	})
}

func RecursiveDFS(node *CM_GraphNode, nodes []CM_GraphNode, edges []CM_Edge, nodeIndexMap map[string]int, groupNum, depth int) {
	// Clamped DFS
	if depth > 2 {
		return
	}
	// Discover it
	node.Group = groupNum
	// Add the neighbors to the stack
	for _, edge := range edges {
		if edge.Source == node.BioGuideId {
			neighbor := &nodes[nodeIndexMap[edge.Target]]
			if neighbor.NotVisited() {
				RecursiveDFS(neighbor, nodes, edges, nodeIndexMap, groupNum, depth+1)
			}
		}
		if edge.Target == node.BioGuideId {
			neighbor := &nodes[nodeIndexMap[edge.Source]]
			if neighbor.NotVisited() {
				RecursiveDFS(neighbor, nodes, edges, nodeIndexMap, groupNum, depth+1)
			}
		}
	}
}

func SetGroupsViaConnectedComponents(nodes []CM_GraphNode, edges []CM_Edge) int {
	groupNum := 1

	fmt.Println("Total Edge Count = ", len(edges))

	// Build a map of the nodes
	visited := make(map[string]int, len(nodes))
	for i, node := range nodes {
		// Reset the group identity
		nodes[i].Group = 0

		// Create a map of the nodes to make it easier to find them
		visited[node.BioGuideId] = i
	}

	first := true
	for _, nodeNum := range visited {
		node := &nodes[nodeNum]
		if first {
			log.Println("Starting at node", node.BioGuideId, node.Group)
			first = false
		}
		if node.NotVisited() {
			fmt.Println("Working on node", node.BioGuideId)
			// Do the DFS
			RecursiveDFS(node, nodes, edges, visited, groupNum, 0)

			// Start a new group
			groupNum++
		}
	}

	x := make(map[int]bool, 500)
	for _, node := range nodes {
		x[node.Group] = true
	}
	log.Println("Found", groupNum-1, len(x), "connected components")

	return groupNum
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

	// What componenet id this node belongs to
	Group int

	// Radius
	R float64
}

func (c CM_GraphNode) NotVisited() bool {
	return c.Group == 0
}
