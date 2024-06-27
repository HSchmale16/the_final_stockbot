package main

import (
	"io"
	"os"
	"testing"
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
		CongressMembers: []CongressMember{
			{Party: "R", State: "TN", BioGuideId: "B001309", Role: "SPONSOR", Name: "Burchett, Tim"},
			{Party: "D", State: "FL", BioGuideId: "M001217", Role: "COSPONSOR", Name: "Moskowitz, Jared"},
			{Party: "R", State: "NY", BioGuideId: "L000599", Role: "COSPONSOR", Name: "Lawler, Michael"},
			{Party: "D", State: "NJ", BioGuideId: "G000583", Role: "COSPONSOR", Name: "Gottheimer, Josh"},
		},
	}

	if !expected.Equals(result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
