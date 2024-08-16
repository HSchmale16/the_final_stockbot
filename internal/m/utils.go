package m

import (
	"log"
	"strings"
)

func SplitName(s string) (string, string) {
	// First convert any accented characters to a percent for LIKE matching
	s = strings.ReplaceAll(s, "á", "%")

	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		log.Printf("Could not split name: %s", s)
		return "", ""
	}
	// trim each
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])

	// log.Printf("Last: %s, First: %s", parts[0], parts[1])
	return parts[0], parts[1]
}

func CollapseTerms(oldTerms []Terms) []Terms {
	// Collapse the terms
	// Terms can be collapsed if they in the same state, district, party, and type.
	// We also collapse if the end year is the same as the start year of the next term.
	var terms []Terms
	terms = append(terms, oldTerms[0])
	for i := 1; i < len(oldTerms); i++ {
		if oldTerms[i].State == oldTerms[i-1].State &&
			oldTerms[i].Party == oldTerms[i-1].Party &&
			oldTerms[i].Type == oldTerms[i-1].Type && oldTerms[i].Start[0:4] == oldTerms[i-1].End[0:4] {
			// Collapse the terms
			terms[len(terms)-1].End = oldTerms[i].End
		} else {
			terms = append(terms, oldTerms[i])
		}
	}
	return terms
}
