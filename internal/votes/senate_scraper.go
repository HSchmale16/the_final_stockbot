package votes

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/travel"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"gorm.io/gorm"
)

func LoadSenateRollCallXml(url string, db *gorm.DB) error {
	bytes, err := senatelobbying.SendRequest(url)
	if err != nil {
		return err
	}

	senateRollCall := ProcessSenateXml(bytes)
	fmt.Println(senateRollCall)

	return ProcessSenateRollCall(url, senateRollCall, db)
}

func ProcessSenateRollCall(url string, senateRollCall SenateRollCall, db *gorm.DB) error {
	vote := Vote{
		Url:         url,
		RollCallNum: senateRollCall.VoteNumber,
		CongressNum: senateRollCall.Congress,
		Session:     senateRollCall.Session,
		Chamber:     "Senate",
		ActionAt:    senateRollCall.VoteDate,
		VoteResult:  senateRollCall.VoteResult,
		LegisName:   senateRollCall.VoteDetails[0],
	}

	err := db.Transaction(func(tx *gorm.DB) error {

		tx2 := tx.Create(&vote)
		if tx2.Error != nil {
			return tx2.Error
		}

		for _, member := range senateRollCall.Members {
			first, last := member.First, member.Last
			first = strings.ToUpper(first)
			last = strings.ToUpper(last)
			senator, err := travel.FuzzyFindSenator(db, last, first, member.State)
			if err != nil {
				return err
			}

			voteRecord := VoteRecord{
				VoteID:     vote.ID,
				MemberId:   senator.BioGuideId,
				VoteStatus: member.VoteCast,
			}

			tx2 = tx.Create(&voteRecord)
			if tx2.Error != nil {
				return tx2.Error
			}

		}

		return nil
	})

	return err
}

func ProcessSenateXml(data []byte) SenateRollCall {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	result := SenateRollCall{}

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "vote_date" {
				// July 11, 2024, 11:54 AM
				var voteDate string
				decoder.DecodeElement(&voteDate, &se)
				result.VoteDate = decodeTime(voteDate)
			}
			if se.Name.Local == "modify_date" {
				var voteModify string
				decoder.DecodeElement(&voteModify, &se)
				result.VoteModify = decodeTime(voteModify)
			}
			if se.Name.Local == "session" {
				var session string
				decoder.DecodeElement(&session, &se)
				result.Session = session
			}
			if se.Name.Local == "congress" {
				var congress int
				decoder.DecodeElement(&congress, &se)
				result.Congress = congress
			}
			if se.Name.Local == "member" {
				var member Member
				decoder.DecodeElement(&member, &se)
				result.Members = append(result.Members, member)
			}
			if se.Name.Local == "vote_number" {
				var voteNumber int
				decoder.DecodeElement(&voteNumber, &se)
				result.VoteNumber = voteNumber
			}
			if se.Name.Local == "vote_result" {
				var voteResult string
				decoder.DecodeElement(&voteResult, &se)
				result.VoteResult = voteResult
			}
			if se.Name.Local == "vote_question_text" || se.Name.Local == "vote_document_text" || se.Name.Local == "vote_title" {
				var voteDetail string
				decoder.DecodeElement(&voteDetail, &se)
				result.VoteDetails = append(result.VoteDetails, voteDetail)
			}
		}

	}

	return result
}

func decodeTime(t string) time.Time {
	// July 11, 2024, 11:54 AM
	layout := "January 2, 2006, 3:04 PM"
	tm, err := time.Parse(layout, t)
	if err != nil {
		panic(err)
	}
	return tm
}

type SenateRollCall struct {
	VoteDate    time.Time
	VoteModify  time.Time
	Congress    int
	Session     string
	VoteNumber  int
	VoteDetails []string

	VoteResult string

	Members []Member
}

type Member struct {
	First    string `xml:"first_name"`
	Last     string `xml:"last_name"`
	Party    string `xml:"party"`
	State    string `xml:"state"`
	VoteCast string `xml:"vote_cast"`
}
