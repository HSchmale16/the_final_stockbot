package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func PieChart(t *testing.T) {
	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: []chart.Value{
			{
				Value: 1,
				Label: "R",
				Style: chart.Style{
					FillColor: drawing.ColorBlue,
				},
			},
		},
		Title: "Which Parties works with",
	}
	b := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.SVG, b)

	assert.Nil(t, err)

	f, err := os.Create("test_pie_chart.svg")
	assert.Nil(t, err)
	defer f.Close()

	_, err = f.Write(b.Bytes())
	assert.Nil(t, err)
}
