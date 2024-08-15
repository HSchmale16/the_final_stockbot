package votes

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"gorm.io/gorm"
)

type HouseRollCallXml struct {
	LegisNum         string
	RollCallNum      int
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
	VoteDesc         string
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
			if se.Name.Local == "rollcall-num" {
				var rollCallNum int
				decoder.DecodeElement(&rollCallNum, &se)
				result.RollCallNum = rollCallNum
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

			if se.Name.Local == "vote-desc" {
				var voteDesc string
				decoder.DecodeElement(&voteDesc, &se)
				result.VoteDesc = voteDesc
			}
		}
	}
	return result
}

func LoadHouseRollCallXml(url string, db *gorm.DB) {
	fmt.Println("Loading", url)
	data, err := senatelobbying.SendRequest(url)
	if err != nil {
		panic(err)
	}
	res := ProcessHouseRollCallXml(data)
	// get the real time instance

	actionAt, err := time.Parse("2-Jan-2006 3:04 PM", res.ActionDate+" "+res.ActionTime)
	if err != nil {
		panic(err)
	}

	vote := Vote{
		RollCallNum:      res.RollCallNum,
		CongressNum:      res.CongressNum,
		Chamber:          res.Chamber,
		ActionAt:         actionAt,
		VoteType:         res.VoteType,
		LegisName:        res.LegisNum,
		VoteResult:       res.Result,
		AmmendmentNum:    res.AmmendmentNum,
		AmmendmentAuthor: res.AmmendmentAuthor,
		Session:          res.Session,
		VoteDesc:         res.VoteDesc,
	}

	fmt.Println("Trying to find ", res.RollCallNum, res.CongressNum, res.Session, res.Chamber)
	x := db.Debug().Where(Vote{
		RollCallNum: res.RollCallNum,
		CongressNum: res.CongressNum,
		Session:     res.Session,
		Chamber:     res.Chamber,
	}).Attrs(vote).FirstOrCreate(&vote)
	if x.Error != nil {
		log.Fatal(x.Error)
	}

	var voteRecords = make([]VoteRecord, len(res.Votes))
	memberIdFixer := map[string]string{
		"L000555": "L000595",
	}

	for i, v := range res.Votes {
		if memberIdFixer[v.Legislator.NameId] != "" {
			v.Legislator.NameId = memberIdFixer[v.Legislator.NameId]
		}

		voteRecords[i] = VoteRecord{
			MemberId:   v.Legislator.NameId,
			VoteStatus: v.Vote,
			VoteID:     vote.ID,
		}
	}

	x = db.Debug().CreateInBatches(&voteRecords, 3)
	if x.Error != nil {
		log.Fatalln(x.Error)
	}

}
