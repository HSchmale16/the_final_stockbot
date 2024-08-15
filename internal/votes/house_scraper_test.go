package votes_test

import (
	_ "embed"
	"testing"

	"github.com/hschmale16/the_final_stockbot/internal/votes"
)

//go:embed test_data/house_2023_roll099.xml
var house_roll_call_xml []byte

func TestProcessHouseRollCall(t *testing.T) {

	res := votes.ProcessHouseRollCallXml(house_roll_call_xml)

	if res.LegisNum != "H R 382" {
		t.Error("LegisNum is not correct")
	}

	if res.CongressNum != 118 {
		t.Error("CongressNum is not correct")
	}

	if res.Session != "1st" {
		t.Error("Session is not correct")
	}

	if res.Chamber != "U.S. House of Representatives" {
		t.Error("Chamber is not correct")
	}

	if res.Question != "On Motion to Recommit" {
		t.Error("Question is not correct")
	}

	if res.Result != "Failed" {
		t.Error("Result is not correct")
	}

	if res.Question != "On Motion to Recommit" {
		t.Error("Question is not correct")
	}

	if res.VoteType != "YEA-AND-NAY" {
		t.Error("VoteType is not correct")
	}

	if res.ActionDate != "31-Jan-2023" {
		t.Error("ActionDate is not correct, got, '", res.ActionDate, "'")
	}

	if res.ActionTime != "5:21 PM" {
		t.Error("ActionTime is not correctt, got '", res.ActionTime, "'")
	}

	if len(res.Votes) != 434 {
		t.Error("Votes length is not correct got ", len(res.Votes))
	}
}
