package congress

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	. "github.com/hschmale16/the_final_stockbot/internal/m"
)

func LoadCongressCommittees(file string) {
	// Open the Database
	db, err := SetupDB()
	if err != nil {
		panic(err)
	}

	// Load the committees from the file
	jsonFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	// Read it
	var committees []JSON_CongressCommittee
	json.Unmarshal(data, &committees)

	fmt.Println("Committees:", len(committees))
	fmt.Println("First Committee:", committees[0])

	// Save it

	for _, committee := range committees {
		// Save the committee
		var myCommittee DB_CongressCommittee
		db.First(&myCommittee, committee.ThomasId)

		myCommittee.F_CongressCommittee = committee.F_CongressCommittee
		myCommittee.ParentCommitteeId = nil
		db.Save(&myCommittee)

		for _, subcommittee := range committee.Subcommittees {
			var mySubcommittee DB_CongressCommittee
			db.First(&mySubcommittee, subcommittee.ThomasId)

			mySubcommittee.F_CongressCommittee = subcommittee
			mySubcommittee.ThomasId = myCommittee.ThomasId + subcommittee.ThomasId
			mySubcommittee.ParentCommitteeId = &myCommittee.ThomasId
			db.Save(&mySubcommittee)
		}
	}
}

func LoadCommitteeMemberships(file string) {
	// Open the Database
	db, err := SetupDB()
	if err != nil {
		panic(err)
	}

	// Load the committees from the file
	jsonFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	var memberships map[string][]json_CommitteeMembership
	json.Unmarshal(data, &memberships)

	fmt.Println("Memberships:", len(memberships))

	for committeeId, committeeMembers := range memberships {
		fmt.Println("Committee:", committeeId, "Members:", len(committeeMembers))

		// Save the committee members
		for _, member := range committeeMembers {
			fmt.Println("Member:", member)

			membership := DB_CommitteeMembership{
				CongressMemberId: member.Bioguide,
				CommitteeId:      committeeId,
				Rank:             member.Rank,
				Title:            member.Title,
				//CongressNumber:   118, // TODO: Will need to update at beginning of 2025 for the 119th Congress
			}
			db.Save(&membership)
		}
	}
}

type json_CommitteeMembership struct {
	Name     string `json:"name"`
	Rank     int    `json:"rank"`
	Bioguide string `json:"bioguide"`
	Title    string `json:"title"`
	Party    string `json:"party"`
}
