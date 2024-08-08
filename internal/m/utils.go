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
