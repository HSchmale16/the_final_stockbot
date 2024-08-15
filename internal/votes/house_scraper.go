package votes

import (
	"bytes"
	"encoding/xml"
	"fmt"

	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"gorm.io/gorm"
)

type HouseRollCallXml struct {
	LegisNum         string
	CongressNum      int
	Session          string
	Chamber          string
	Question         string
	Result           string
	VoteType         string
	ActionDate       string
	ActionTime       string
	AmmendmentNum    int
	AmmendmentAuthor string
	Votes            []recordedVote
}

type recordedVote struct {
	Legislator struct {
		NameId string `xml:"name-id,attr"`
	} `xml:"legislator"`
	Vote string `xml:"vote"`
}

func ProcessHouseRollCallXml(data []byte) HouseRollCallXml {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	result := HouseRollCallXml{}

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "legis-num" {
				var legisNum string
				decoder.DecodeElement(&legisNum, &se)
				result.LegisNum = legisNum
			}
			if se.Name.Local == "recorded-vote" {
				var vote recordedVote
				decoder.DecodeElement(&vote, &se)
				result.Votes = append(result.Votes, vote)
			}
			if se.Name.Local == "congress" {
				var congressNum int
				decoder.DecodeElement(&congressNum, &se)
				result.CongressNum = congressNum
			}
			if se.Name.Local == "session" {
				var session string
				decoder.DecodeElement(&session, &se)
				result.Session = session
			}
			if se.Name.Local == "chamber" {
				var chamber string
				decoder.DecodeElement(&chamber, &se)
				result.Chamber = chamber
			}
			if se.Name.Local == "vote-question" {
				var question string
				decoder.DecodeElement(&question, &se)
				result.Question = question
			}
			if se.Name.Local == "vote-result" {
				var voteResult string
				decoder.DecodeElement(&voteResult, &se)
				result.Result = voteResult
			}
			if se.Name.Local == "vote-type" {
				var voteType string
				decoder.DecodeElement(&voteType, &se)
				result.VoteType = voteType
			}
			if se.Name.Local == "action-date" {
				var actionDate string
				decoder.DecodeElement(&actionDate, &se)
				result.ActionDate = actionDate
			}
			if se.Name.Local == "action-time" {
				var actionTime string
				decoder.DecodeElement(&actionTime, &se)
				result.ActionTime = actionTime
			}
			if se.Name.Local == "amendment-num" {
				var ammendmentNum int
				decoder.DecodeElement(&ammendmentNum, &se)
				result.AmmendmentNum = ammendmentNum
			}
			if se.Name.Local == "amendment-author" {
				var ammendmentAuthor string
				decoder.DecodeElement(&ammendmentAuthor, &se)
				result.AmmendmentAuthor = ammendmentAuthor
			}
		}
	}
	return result
}

func LoadHouseRollCallXml(url string, db *gorm.DB) {
	data, err := senatelobbying.SendRequest(url)
	if err != nil {
		panic(err)
	}
	res := ProcessHouseRollCallXml(data)
	fmt.Println(res)
}
