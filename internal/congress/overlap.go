package congress

import (
	"github.com/ernestosuarez/itertools"
)

type GroupOverlaps struct {
	Groups []string `json:"sets"`
	Label  string   `json:"label"`
	Count  int      `json:"size"`
}

func computeOverlap(committees []DB_CongressCommittee) []GroupOverlaps {
	committeeIds2Members := make(map[string][]DB_CommitteeMembership)
	committeeIds := make([]string, 1)

	for _, committee := range committees {
		committeeIds = append(committeeIds, committee.ThomasId)
		committeeIds2Members[committee.ThomasId] = committee.Memberships
	}

	overlaps := make([]GroupOverlaps, 0, 3*len(committees))

	for combination := range itertools.CombinationsStr(committeeIds, 2) {
		count := 0
		// if there is an empty string in combination, it is a combination of 1
		if combination[0] == "" {
			count = len(committeeIds2Members[combination[1]])
			combination = []string{combination[1]}
		} else if combination[1] == "" {
			count = len(committeeIds2Members[combination[0]])
			combination = []string{combination[0]}
		} else {
			for _, member1 := range committeeIds2Members[combination[0]] {
				for _, member2 := range committeeIds2Members[combination[1]] {
					if member1.CongressMemberId == member2.CongressMemberId {
						count++
					}
				}
			}
		}

		var label string
		if len(combination) == 1 {
			// Find the right label linear search that thing
			for _, committee := range committees {
				if committee.ThomasId == combination[0] {
					label = committee.Name
					break
				}
			}
		}

		overlaps = append(overlaps, GroupOverlaps{
			Groups: combination,
			Label:  label,
			Count:  int(count),
		})
	}

	return overlaps
}
