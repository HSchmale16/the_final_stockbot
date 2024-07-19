/**
 * Reads the mods XML data and generates a structure containing details about the data
 */

package m

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
)

type XML_CongressCommittee struct {
	Name        string `xml:"name"`
	Chamber     string `xml:"chamber,attr"`
	AuthorityId string `xml:"authorityId,attr"`
}

type CongressMember struct {
	Party      string `xml:"party,attr"`
	State      string `xml:"state,attr"`
	BioGuideId string `xml:"bioGuideId,attr"`
	Role       string `xml:"role,attr"`
	Name       string `xml:"name"`
	Chamber    string `xml:"chamber,attr"`
	Congress   string `xml:"congress,attr"`
}

type ModsAction struct {
	Date string `xml:"date,attr"`
	Text string `xml:",chardata"`
}

type LawModsData struct {
	OfficialTitle      string
	Actions            []ModsAction
	CongressCommittees []XML_CongressCommittee
	CongressMembers    []CongressMember
}

func (l *LawModsData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan US_CongressLegislators")
	}
	return json.Unmarshal(b, l)
}

func (l LawModsData) Value() (driver.Value, error) {
	return json.Marshal(l)

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
				var committee XML_CongressCommittee
				decoder.DecodeElement(&committee, &se)
				modsData.CongressCommittees = append(modsData.CongressCommittees, committee)
			}
			if se.Name.Local == "action" {
				var action ModsAction
				decoder.DecodeElement(&action, &se)
				modsData.Actions = append(modsData.Actions, action)
			}
			if se.Name.Local == "officialTitle" {
				// read char data from element
				decoder.DecodeElement(&modsData.OfficialTitle, &se)
			}

		}
	}
	return modsData
}
