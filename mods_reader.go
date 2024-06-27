/**
 * Reads the mods XML data and generates a structure containing details about the data
 */

package main

import (
	"encoding/xml"
	"strings"
)

type CongressCommittee struct {
	Name    string `xml:"name"`
	Chamber string `xml:"chamber,attr"`
}

type CongressMember struct {
	Party      string `xml:"party,attr"`
	State      string `xml:"state,attr"`
	BioGuideId string `xml:"bioGuideId,attr"`
	Role       string `xml:"role,attr"`
	Name       string `xml:"name"`
}

type LawModsData struct {
	CongressCommittees []CongressCommittee
	CongressMembers    []CongressMember
}

func (l LawModsData) Equals(other LawModsData) bool {
	if len(l.CongressMembers) != len(other.CongressMembers) {
		return false
	}

	for i := range l.CongressMembers {
		if l.CongressMembers[i] != other.CongressMembers[i] {
			return false
		}
	}

	return true
}

func ReadLawModsData(xmlString string) LawModsData {
	decoder := xml.NewDecoder(strings.NewReader(xmlString))
	modsData := LawModsData{}

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "congMember" {
				var member CongressMember
				decoder.DecodeElement(&member, &se)
				modsData.CongressMembers = append(modsData.CongressMembers, member)
			}
			if se.Name.Local == "congCommittee" {
				var committee CongressCommittee
				decoder.DecodeElement(&committee, &se)
				modsData.CongressCommittees = append(modsData.CongressCommittees, committee)
			}
		}
	}
	return modsData
}
