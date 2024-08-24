package app_test

import (
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
	var edges = []app.CM_Edge{
		{
			Source: "A",
			Target: "B",
		},
		{
			Source: "F",
			Target: "G",
		},
		{
			Source: "B",
			Target: "C",
		},
		{
			Source: "D",
			Target: "E",
		},
		{
			Source: "W",
			Target: "X",
		},
		{
			Source: "X",
			Target: "Y",
		},
		{
			Source: "Y",
			Target: "X",
		},
		{
			Source: "Z",
			Target: "X",
		},
	}

	app.SetGroupsViaConnectedComponents(nodes, edges)
	// Double check the numbers
	x := make(map[int]bool, 500)
	for _, node := range nodes {
		x[node.Group] = true
	}

	t.Log("Found", len(x), "connected components")

	t.Fail()
}
