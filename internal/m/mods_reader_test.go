package m_test

import (
	"io"
	"os"
	"testing"

	. "github.com/hschmale16/the_final_stockbot/internal/m"
)

func TestModsXML(t *testing.T) {
	file, err := os.Open("testData/mod1.xml")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Perform your test logic using the file

	// Example: Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	result := ReadLawModsData(string(data))

	expected := LawModsData{
		OfficialTitle: "Rejecting the United Nations decision to place the Israel Defense Force on a list of child’s rights abusers.",
		Actions: []ModsAction{{
			Date: "2024-06-26",
			Text: "Mr. Burchett (for himself, Mr. Moskowitz, Mr. Lawler, and Mr. Gottheimer) submitted the following resolution; which was referred to the Committee on Foreign Affairs",
		}},
		CongressCommittees: []XML_CongressCommittee{{
			Name:    "Foreign Affairs",
			Chamber: "H",
		}},
		CongressMembers: []CongressMember{
			{Party: "R", State: "TN", BioGuideId: "B001309", Role: "SPONSOR", Name: "Burchett, Tim", Chamber: "H", Congress: "118"},
			{Party: "D", State: "FL", BioGuideId: "M001217", Role: "COSPONSOR", Name: "Moskowitz, Jared", Chamber: "H", Congress: "118"},
			{Party: "R", State: "NY", BioGuideId: "L000599", Role: "COSPONSOR", Name: "Lawler, Michael", Chamber: "H", Congress: "118"},
			{Party: "D", State: "NJ", BioGuideId: "G000583", Role: "COSPONSOR", Name: "Gottheimer, Josh", Chamber: "H", Congress: "118"},
		},
	}

	if !expected.Equals(result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
