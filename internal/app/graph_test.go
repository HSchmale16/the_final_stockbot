package app_test

import (
	"fmt"
	"testing"

	"github.com/hschmale16/the_final_stockbot/internal/app"
)

func TestCongressNetwork(t *testing.T) {
	var nodes = []app.CM_GraphNode{}
	// Create 26 Nodes from A to Z
	for i := 0; i < 26; i++ {
		nodes = append(nodes, app.CM_GraphNode{
			BioGuideId: string(rune('A' + i)),
		})
	}
	var edges = []app.CM_Edge{}
	// Create 25 edges from A to Z
	for i := 0; i < 25; i++ {
		edges = append(edges, app.CM_Edge{
			Source: string(rune('A' + i)),
			Target: string(rune('A' + i + 1)),
		})
	}
	fmt.Println(edges)

	groupNum := app.SetGroupsViaConnectedComponents(nodes, edges)
	// Double check the numbers
	// Double check the numbers
	x := make(map[int]bool, 500)
	for _, node := range nodes {
		x[node.Group] = true
		if node.Group == 0 {
			t.Error("Group not set -> ", node.BioGuideId)
		}
	}
	// check for the number of connected components
	for i := 1; i < groupNum; i++ {
		if !x[i] {
			t.Error("Missing group", i)
		}
	}

	if len(x) != 1 {
		t.Error("Expected 1 group, got", groupNum)
	}
}

func TestNotVisited(t *testing.T) {
	node := app.CM_GraphNode{
		Group: 0,
	}

	if !node.NotVisited() {
		t.Error("Node is visited")
	}

	node.Group = 1
	if node.NotVisited() {
		t.Error("Node is not visited")
	}
}
